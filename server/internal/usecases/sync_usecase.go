package usecases

import (
	"fmt"
	"log"

	"chantry/server/internal/discord"
	"chantry/server/internal/pocketbase"
)

// Metrics consolidates the results of the logical upsert execution
type Metrics struct {
	TotalProcessed int `json:"total_processed"`
	NewInserted    int `json:"new_inserted"`
	Updated        int `json:"updated"`
}

// StudentSyncPayload defines the input structure for student role sync
type StudentSyncPayload struct {
	PrimaryRoleID    string   `json:"primary_role_id"`
	SecondaryRoleIDs []string `json:"secondary_role_ids"`
}

// ManagerSyncRule maps a Discord role to a pre-defined manager role access level
type ManagerSyncRule struct {
	RoleID      string `json:"role_id"`
	ManagerType string `json:"manager_type"` // admin, mentor, pedagogy
}

// AdvancedSyncPayload holds parameters for the multi-role student and manager sync
type AdvancedSyncPayload struct {
	Students StudentSyncPayload `json:"students"`
	Managers []ManagerSyncRule  `json:"managers"`
}

// AdvancedSyncMetrics holds mutation summaries for students and managers
type AdvancedSyncMetrics struct {
	StudentsProcessed int `json:"students_processed"`
	StudentsInserted  int `json:"students_inserted"`
	StudentsUpdated   int `json:"students_updated"`
	ManagersProcessed int `json:"managers_processed"`
	ManagersInserted  int `json:"managers_inserted"`
	ManagersUpdated   int `json:"managers_updated"`
}

// SyncUsecase orchestrates the logic to fetch, validate and persist Discord members into PocketBase
type SyncUsecase struct {
	discordService   *discord.DiscordService
	pbRepo           *pocketbase.Repository
	provisionUsecase *ProvisionUsecase
}

// NewSyncUsecase creates a new instance of SyncUsecase
func NewSyncUsecase(discordService *discord.DiscordService, pbRepo *pocketbase.Repository, provisionUsecase *ProvisionUsecase) *SyncUsecase {
	return &SyncUsecase{
		discordService:   discordService,
		pbRepo:           pbRepo,
		provisionUsecase: provisionUsecase,
	}
}

// EnsureRoleExists resolves or creates a Role in PocketBase, returning its 15-character ID
func (u *SyncUsecase) EnsureRoleExists(guildID, roleID, pbGuildID string) (string, error) {
	var roleRecord pocketbase.RoleRecord
	found, err := u.pbRepo.FindFirstByDiscordID("roles", roleID, &roleRecord)
	if err != nil {
		return "", err
	}

	if !found {
		roleName := "Sincronizado via Discord"
		if dRoles, err := u.discordService.Session.GuildRoles(guildID); err == nil {
			for _, r := range dRoles {
				if r.ID == roleID {
					roleName = r.Name
					break
				}
			}
		} else {
			log.Printf("⚠️ Warning: Failed to fetch Roles for Guild %s from Discord API. Error: %v", guildID, err)
		}

		newRole := pocketbase.RoleRecord{
			DiscordID: roleID,
			Name:      roleName,
			GuildID:   pbGuildID,
		}
		if err := u.pbRepo.CreateRecord("roles", newRole, &roleRecord); err != nil {
			return "", err
		}
		log.Printf("✅ Auto-created missing Role record: %s (PB ID: %s)", roleName, roleRecord.ID)
	}

	return roleRecord.ID, nil
}

// SyncStudentsByRole performs cursor-paginated member searches in Discord and syncs them to PocketBase
func (u *SyncUsecase) SyncStudentsByRole(guildID string, roleID string) (Metrics, error) {
	metrics := Metrics{}

	// 1. Ensure Guild exists in PocketBase (Silent Upsert)
	var guildRecord pocketbase.GuildRecord
	found, err := u.pbRepo.FindFirstByDiscordID("guilds", guildID, &guildRecord)
	if err != nil {
		return metrics, fmt.Errorf("failed to query guild in pocketbase: %w", err)
	}

	if !found {
		guildName := "Sincronizado via Discord"
		// Query Discord REST API for Guild information
		if dGuild, err := u.discordService.Session.Guild(guildID); err == nil && dGuild != nil {
			guildName = dGuild.Name
		} else {
			log.Printf("⚠️ Warning: Failed to fetch Guild %s details from Discord API, using fallback name. Error: %v", guildID, err)
		}

		newGuild := pocketbase.GuildRecord{
			DiscordID: guildID,
			Name:      guildName,
			Status:    "active",
		}
		if err := u.pbRepo.CreateRecord("guilds", newGuild, &guildRecord); err != nil {
			return metrics, fmt.Errorf("failed to auto-create guild record in pocketbase: %w", err)
		}
		log.Printf("✅ Auto-created missing Guild record: %s (PB ID: %s)", guildName, guildRecord.ID)
	}

	// 2. Ensure Role exists in PocketBase (Silent Upsert)
	var roleRecord pocketbase.RoleRecord
	found, err = u.pbRepo.FindFirstByDiscordID("roles", roleID, &roleRecord)
	if err != nil {
		return metrics, fmt.Errorf("failed to query role in pocketbase: %w", err)
	}

	if !found {
		roleName := "Sincronizado via Discord"
		// Query Discord REST API to fetch Guild Roles
		if dRoles, err := u.discordService.Session.GuildRoles(guildID); err == nil {
			for _, r := range dRoles {
				if r.ID == roleID {
					roleName = r.Name
					break
				}
			}
		} else {
			log.Printf("⚠️ Warning: Failed to fetch Roles for Guild %s from Discord API, using fallback name. Error: %v", guildID, err)
		}

		newRole := pocketbase.RoleRecord{
			DiscordID: roleID,
			Name:      roleName,
			GuildID:   guildRecord.ID, // PocketBase 15-char record ID relation
		}
		if err := u.pbRepo.CreateRecord("roles", newRole, &roleRecord); err != nil {
			return metrics, fmt.Errorf("failed to auto-create role record in pocketbase: %w", err)
		}
		log.Printf("✅ Auto-created missing Role record: %s (PB ID: %s)", roleName, roleRecord.ID)
	}

	// 3. Fetch active Discord guild members holding the role
	members, err := u.discordService.GetGuildMembersByRole(guildID, roleID)
	if err != nil {
		return metrics, fmt.Errorf("failed to fetch guild members by role from Discord API: %w", err)
	}

	// 4. Perform logical upsert for each student member
	for _, m := range members {
		metrics.TotalProcessed++

		var student pocketbase.StudentRecord
		studentFound, err := u.pbRepo.FindFirstByDiscordAndGuild("students", m.ID, guildRecord.ID, &student)
		if err != nil {
			log.Printf("⚠️ Error querying student %s (%s) in guild %s: %v", m.Username, m.ID, guildRecord.ID, err)
			continue
		}

		if studentFound {
			// Compare dynamic data to check if partial PATCH is required (optimizing API requests)
			if student.Username != m.Username || student.Nickname != m.Nickname || student.RoleID != roleRecord.ID || student.GuildID != guildRecord.ID {
				updateData := map[string]interface{}{
					"username":  m.Username,
					"nickname":  m.Nickname,
					"role_id":   roleRecord.ID,
					"guild_id":  guildRecord.ID,
				}
				var updatedStudent pocketbase.StudentRecord
				if err := u.pbRepo.UpdateRecord("students", student.ID, updateData, &updatedStudent); err != nil {
					log.Printf("⚠️ Error updating student record %s (ID: %s) in PocketBase: %v", m.Username, student.ID, err)
					continue
				}
				metrics.Updated++
			}
		} else {
			// Create a brand new active student record
			newStudent := pocketbase.StudentRecord{
				DiscordID: m.ID,
				Username:  m.Username,
				Nickname:  m.Nickname,
				RoleID:    roleRecord.ID,
				GuildID:   guildRecord.ID,
				Status:    "active",
			}
			var createdStudent pocketbase.StudentRecord
			if err := u.pbRepo.CreateRecord("students", newStudent, &createdStudent); err != nil {
				log.Printf("⚠️ Error creating student record %s in PocketBase: %v", m.Username, err)
				continue
			}
			metrics.NewInserted++
		}
	}

	return metrics, nil
}

// AdvancedSync performs multi-role students synchronization using the guild's taxonomy and triggers auto-healing.
func (u *SyncUsecase) AdvancedSync(guildID string) (AdvancedSyncMetrics, error) {
	metrics := AdvancedSyncMetrics{}

	// 1. Ensure Guild exists in PocketBase (Silent Upsert)
	var guildRecord pocketbase.GuildRecord
	found, err := u.pbRepo.FindFirstByDiscordID("guilds", guildID, &guildRecord)
	if err != nil {
		return metrics, fmt.Errorf("failed to query guild in pocketbase: %w", err)
	}

	if !found {
		guildName := "Sincronizado via Discord"
		if dGuild, err := u.discordService.Session.Guild(guildID); err == nil && dGuild != nil {
			guildName = dGuild.Name
		}
		newGuild := pocketbase.GuildRecord{
			DiscordID: guildID,
			Name:      guildName,
			Status:    "active",
		}
		if err := u.pbRepo.CreateRecord("guilds", newGuild, &guildRecord); err != nil {
			return metrics, fmt.Errorf("failed to auto-create guild record in pocketbase: %w", err)
		}
	}

	// 2. Perform Students Sync iterating over all SquadRoles defined in Taxonomy
	if len(guildRecord.SquadRoles) > 0 {
		for _, squadRoleID := range guildRecord.SquadRoles {
			log.Printf("🚀 [ADV-SYNC] Syncing students for Squad Role %s", squadRoleID)
			
			// A. Executa a sincronização base de membros (Trazendo do Discord para o PocketBase)
			roleMetrics, err := u.SyncStudentsByRole(guildID, squadRoleID)
			if err != nil {
				log.Printf("❌ ERROR [ADV-SYNC] Failed to sync students for role %s: %v", squadRoleID, err)
				continue
			}

			// Acumulando as métricas baseadas no retorno
			metrics.StudentsProcessed += roleMetrics.TotalProcessed
			metrics.StudentsInserted += roleMetrics.NewInserted
			metrics.StudentsUpdated += roleMetrics.Updated

			// B. Localizar o PB ID do cargo para verificar se há SquadChannelID (Categoria 1-on-1)
			var roleRecord pocketbase.RoleRecord
			roleFound, err := u.pbRepo.FindFirstByDiscordID("roles", squadRoleID, &roleRecord)
			if err != nil {
				log.Printf("⚠️ Warning: Failed to query role %s in pocketbase after sync: %v", squadRoleID, err)
				continue
			}

			if roleFound && roleRecord.SquadChannelID != "" {
				log.Printf("🛠️ [ADV-SYNC] Squad Role %s has associated Category %s. Triggering Auto-Heal...", squadRoleID, roleRecord.SquadChannelID)
				
				// Disparar o Auto Heal (ProvisionUsecase)
				healMetrics, err := u.provisionUsecase.HealChannelsByCategory(guildID, roleRecord.SquadChannelID)
				if err != nil {
					log.Printf("❌ ERROR [ADV-SYNC] Failed to auto-heal channels for Category %s: %v", roleRecord.SquadChannelID, err)
				} else {
					log.Printf("✅ [ADV-SYNC] Auto-Heal completed for Category %s. Mapped: %d, Unmapped: %d", roleRecord.SquadChannelID, healMetrics.SuccessfullyMapped, healMetrics.UnmappedChannels)
				}
			}
		}
	} else {
		log.Printf("⚠️ [ADV-SYNC] Guild %s has no SquadRoles mapped in its taxonomy. Skipping student sync.", guildID)
	}

	return metrics, nil
}


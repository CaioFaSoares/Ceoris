package dto

import "chantry/server/internal/pocketbase"

// StudentListResponse represents the flattened and enriched data structure for Nuxt.
type StudentListResponse struct {
	ID            string `json:"id"`
	DiscordID     string `json:"discord_id"`
	Username      string `json:"username"`
	Nickname      string `json:"nickname"`
	RoleName      string `json:"role_name"`
	GuildID       string `json:"guild_id"`
	Status        string `json:"status"`
	Shift         string `json:"shift"`
	ChannelID     string `json:"channel_id"` // Legacy support for members.vue
	Has1on1       bool   `json:"has_1on1"`
	ChannelStatus string `json:"channel_status"` // "ready" or "pending"
}

// ToStudentListResponse maps a raw PocketBase StudentRecord into a flattened DTO.
func ToStudentListResponse(model pocketbase.StudentRecord) StudentListResponse {
	has1on1 := model.ChannelID != ""
	channelStatus := "pending"
	if has1on1 {
		channelStatus = "ready"
	}

	roleName := model.Expand.Role.Name
	if roleName == "" {
		roleName = "Sem Turma"
	}

	return StudentListResponse{
		ID:            model.ID,
		DiscordID:     model.DiscordID,
		Username:      model.Username,
		Nickname:      model.Nickname,
		RoleName:      roleName,
		GuildID:       model.GuildID,
		Status:        model.Status,
		Shift:         model.Shift,
		ChannelID:     model.ChannelID,
		Has1on1:       has1on1,
		ChannelStatus: channelStatus,
	}
}

package dto

import "chantry/server/internal/pocketbase"

// GuildMappingResponse represents the taxonomy structure for a Guild.
type GuildMappingResponse struct {
	SquadRoles  []string `json:"squad_roles"`
	MentorRoles []string `json:"mentor_roles"`
	SkillRoles  []string `json:"skill_roles"`
}

// ToGuildMappingResponse maps a raw PocketBase GuildRecord into a taxonomy DTO.
func ToGuildMappingResponse(model pocketbase.GuildRecord) GuildMappingResponse {
	mapping := GuildMappingResponse{
		SquadRoles:  model.SquadRoles,
		MentorRoles: model.MentorRoles,
		SkillRoles:  model.SkillRoles,
	}

	// Ensure arrays are not nil
	if mapping.SquadRoles == nil {
		mapping.SquadRoles = []string{}
	}
	if mapping.MentorRoles == nil {
		mapping.MentorRoles = []string{}
	}
	if mapping.SkillRoles == nil {
		mapping.SkillRoles = []string{}
	}

	return mapping
}

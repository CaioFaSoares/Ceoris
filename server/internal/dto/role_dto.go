package dto

import "chantry/server/internal/pocketbase"

// RoleConfigResponse represents the clean data structure for Nuxt.
type RoleConfigResponse struct {
	ID               string `json:"id"`
	DiscordID        string `json:"discord_id"`
	Name             string `json:"name"`
	Shift            string `json:"shift"`
	CheckInTime      string `json:"check_in_time"`
	CheckoutCooldown int    `json:"checkout_cooldown"`
	IsMonitored      bool   `json:"is_monitored"`
	IsActive         bool   `json:"is_active"`
	IsStaff          bool   `json:"is_staff"`
	SquadChannelID   string `json:"squad_channel_id"`
}

// ToRoleConfigResponse maps a raw PocketBase RoleRecord into a flattened DTO.
func ToRoleConfigResponse(model pocketbase.RoleRecord) RoleConfigResponse {
	return RoleConfigResponse{
		ID:               model.ID,
		DiscordID:        model.DiscordID,
		Name:             model.Name,
		Shift:            model.Shift,
		CheckInTime:      model.CheckInTime,
		CheckoutCooldown: model.CheckoutCooldown,
		IsMonitored:      model.IsMonitored,
		IsActive:         model.IsActive,
		IsStaff:          model.IsStaff,
		SquadChannelID:   model.SquadChannelID,
	}
}

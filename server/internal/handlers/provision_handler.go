package handlers

import (
	"log"

	"chantry/server/internal/usecases"
	"github.com/gofiber/fiber/v2"

	"chantry/server/internal/utils"
)

type ProvisionHandler struct {
	provisionUsecase *usecases.ProvisionUsecase
}

// NewProvisionHandler instantiates a new ProvisionHandler.
func NewProvisionHandler(usecase *usecases.ProvisionUsecase) *ProvisionHandler {
	return &ProvisionHandler{
		provisionUsecase: usecase,
	}
}

// HandleProvisionChannels invokes the ProvisionUsecase in background for the entire guild.
func (h *ProvisionHandler) HandleProvisionChannels(c *fiber.Ctx) error {
	guildID := c.Params("id")
	if guildID == "" {
		return utils.JSONError(c, fiber.StatusBadRequest, "The guildId path parameter is required in the route")
	}

	log.Printf("[PROVISION-API] Triggered async batch provisioning for Guild: %s", guildID)
	
	// Fire-and-Forget goroutine
	go func(gID string) {
		log.Printf("[BACKGROUND-PROVISION] Starting batch creation for Guild: %s", gID)
		_, err := h.provisionUsecase.BatchCreatePrivateChannels(gID)
		if err != nil {
			log.Printf("❌ ERROR [BACKGROUND-PROVISION] for Guild ID %s: %v", gID, err)
			return
		}
		log.Printf("✅ SUCCESS [BACKGROUND-PROVISION] for Guild ID %s completed", gID)
	}(guildID)

	return utils.JSONSuccess(c, fiber.StatusAccepted, fiber.Map{
		"status":  "processing",
		"message": "A criação de canais foi iniciada em background",
	})
}

// HandleHealChannels invokes the Usecase to heal student channel mapping in background for the entire guild.
func (h *ProvisionHandler) HandleHealChannels(c *fiber.Ctx) error {
	guildID := c.Params("id")
	if guildID == "" {
		return utils.JSONError(c, fiber.StatusBadRequest, "The guildId path parameter is required in the route")
	}

	log.Printf("[PROVISION-API] Triggered async channel Auto-Healing for Guild: %s", guildID)
	
	// Fire-and-Forget goroutine
	go func(gID string) {
		log.Printf("[BACKGROUND-HEAL] Starting auto-healing for Guild: %s", gID)
		_, err := h.provisionUsecase.HealChannelsByGuild(gID)
		if err != nil {
			log.Printf("❌ ERROR [BACKGROUND-HEAL] for Guild ID %s: %v", gID, err)
			return
		}
		log.Printf("✅ SUCCESS [BACKGROUND-HEAL] for Guild ID %s completed", gID)
	}(guildID)

	return utils.JSONSuccess(c, fiber.StatusAccepted, fiber.Map{
		"status":  "processing",
		"message": "O processo de auto-healing foi iniciado em background",
	})
}

// HandleGetProvisionPageData handles GET /api/ui/provision-page/:guildId, returning the aggregated data for the page.
func (h *ProvisionHandler) HandleGetProvisionPageData(c *fiber.Ctx) error {
	guildID := c.Params("id")
	if guildID == "" {
		return utils.JSONError(c, fiber.StatusBadRequest, "The guildId path parameter is required in the route")
	}

	data, err := h.provisionUsecase.GetProvisionPageData(guildID)
	if err != nil {
		log.Printf("❌ ERROR [HandleGetProvisionPageData] for Guild ID %s: %v", guildID, err)
		return utils.JSONError(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.JSONSuccess(c, fiber.StatusOK, data)
}

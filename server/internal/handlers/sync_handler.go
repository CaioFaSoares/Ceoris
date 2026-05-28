package handlers

import (
	"log"

	"chantry/server/internal/usecases"
	"github.com/gofiber/fiber/v2"

	"chantry/server/internal/utils"
)

// SyncHandler coordinates incoming synchronization HTTP requests
type SyncHandler struct {
	syncUsecase *usecases.SyncUsecase
}

// NewSyncHandler creates a new instance of SyncHandler
func NewSyncHandler(usecase *usecases.SyncUsecase) *SyncHandler {
	return &SyncHandler{
		syncUsecase: usecase,
	}
}

// SyncMembersRequest maps the incoming POST request body
type SyncMembersRequest struct {
	RoleID string `json:"role_id"`
}

// HandleSyncMembers parses inputs, triggers the synchronization usecase, and returns the mutational metrics
func (h *SyncHandler) HandleSyncMembers(c *fiber.Ctx) error {
	guildID := c.Params("id")
	if guildID == "" {
		return utils.JSONError(c, fiber.StatusBadRequest, "The guildId path parameter is required in the route")
	}

	var req SyncMembersRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.JSONError(c, fiber.StatusBadRequest, "Invalid or malformed JSON request body")
	}

	if req.RoleID == "" {
		return utils.JSONError(c, fiber.StatusBadRequest, "The role_id field is required in the JSON body")
	}

	log.Printf("[SYNC] Triggered async member synchronization for Guild: %s, Role: %s", guildID, req.RoleID)
	
	// Fire-and-Forget goroutine
	go func(gID, rID string) {
		log.Printf("[BACKGROUND-SYNC] Starting sync for Guild: %s, Role: %s", gID, rID)
		_, err := h.syncUsecase.SyncStudentsByRole(gID, rID)
		if err != nil {
			log.Printf("❌ ERROR [BACKGROUND-SYNC] for Guild ID %s and Role ID %s: %v", gID, rID, err)
			return
		}
		log.Printf("✅ SUCCESS [BACKGROUND-SYNC] for Guild ID %s and Role ID %s completed", gID, rID)
	}(guildID, req.RoleID)

	return utils.JSONSuccess(c, fiber.StatusAccepted, fiber.Map{
		"status":  "processing",
		"message": "A sincronização de membros foi iniciada em background",
	})
}

// HandleAdvancedSync handles the advanced multi-role and manager sync request using PocketBase Taxonomy
func (h *SyncHandler) HandleAdvancedSync(c *fiber.Ctx) error {
	guildID := c.Params("id")
	if guildID == "" {
		return utils.JSONError(c, fiber.StatusBadRequest, "The guildId path parameter is required in the route")
	}

	log.Printf("[ADV-SYNC] Triggered async advanced synchronization for Guild: %s", guildID)
	
	// Fire-and-Forget goroutine
	go func(gID string) {
		log.Printf("[BACKGROUND-ADV-SYNC] Starting advanced sync for Guild: %s", gID)
		_, err := h.syncUsecase.AdvancedSync(gID)
		if err != nil {
			log.Printf("❌ ERROR [BACKGROUND-ADV-SYNC] for Guild ID %s: %v", gID, err)
			return
		}
		log.Printf("✅ SUCCESS [BACKGROUND-ADV-SYNC] for Guild ID %s completed", gID)
	}(guildID)

	return utils.JSONSuccess(c, fiber.StatusAccepted, fiber.Map{
		"status":  "processing",
		"message": "A sincronização avançada foi iniciada em background",
	})
}

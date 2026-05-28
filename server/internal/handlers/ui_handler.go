package handlers

import (
	"log"

	"chantry/server/internal/usecases"
	"github.com/gofiber/fiber/v2"

	"chantry/server/internal/utils"
)

// UIHandler coordinates incoming HTTP requests specifically meant for the BFF (Streamlit)
type UIHandler struct {
	uiUsecase *usecases.UIUsecase
}

// NewUIHandler instantiates a new HTTP controller for BFF endpoints
func NewUIHandler(uiUsecase *usecases.UIUsecase) *UIHandler {
	return &UIHandler{
		uiUsecase: uiUsecase,
	}
}

// HandleSquadDashboard aggregates data for the Squad Management screen
func (h *UIHandler) HandleSquadDashboard(c *fiber.Ctx) error {
	roleID := c.Params("roleId")
	if roleID == "" {
		return utils.JSONError(c, fiber.StatusBadRequest, "The roleId path parameter is required")
	}

	log.Printf("[UI BFF] Generating squad dashboard data for Role: %s", roleID)

	data, err := h.uiUsecase.GetSquadDashboardData(roleID)
	if err != nil {
		log.Printf("❌ ERROR [HandleSquadDashboard]: %v", err)
		return utils.JSONError(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.JSONSuccess(c, fiber.StatusOK, data)
}

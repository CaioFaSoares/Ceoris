package handlers

import (
	"log"

	"chantry/server/internal/pocketbase"
	"chantry/server/internal/usecases"
	"github.com/gofiber/fiber/v2"

	"chantry/server/internal/utils"
)

type SendBroadcastRequest struct {
	Content    string   `json:"content"`
	TargetType string   `json:"target_type"` // public, private
	RoleIDs    []string `json:"role_ids"`    // opcional
}

type BroadcastHandler struct {
	broadcastUsecase *usecases.BroadcastUsecase
}

func NewBroadcastHandler(usecase *usecases.BroadcastUsecase) *BroadcastHandler {
	return &BroadcastHandler{
		broadcastUsecase: usecase,
	}
}

func (h *BroadcastHandler) HandleSendBroadcast(c *fiber.Ctx) error {
	guildID := c.Params("id")
	if guildID == "" {
		return utils.JSONError(c, fiber.StatusBadRequest, "O parâmetro de rota guildId é obrigatório")
	}

	var req SendBroadcastRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.JSONError(c, fiber.StatusBadRequest, "Corpo da requisição JSON inválido ou malformado")
	}

	// Validações básicas de negócio
	if req.Content == "" {
		return utils.JSONError(c, fiber.StatusBadRequest, "O conteúdo da mensagem (content) não pode ser vazio")
	}

	if req.TargetType != "public" && req.TargetType != "private" {
		return utils.JSONError(c, fiber.StatusBadRequest, "O tipo de destino (target_type) deve ser 'public' ou 'private'")
	}

	log.Printf("[BROADCAST] Recebida solicitação de disparo para a guilda: %s, tipo: %s", guildID, req.TargetType)

	// Chamar Usecase de disparo
	metrics, err := h.broadcastUsecase.SendBroadcast(guildID, req.Content, req.TargetType, req.RoleIDs)
	if err != nil {
		log.Printf("❌ ERRO [HandleSendBroadcast] na guilda %s: %v", guildID, err)

		// Retornar Bad Request caso o erro seja de configuração pendente
		if err.Error() == "canal de avisos não configurado" {
			return utils.JSONError(c, fiber.StatusBadRequest, "Canal de avisos não configurado para este servidor. Configure nas opções de infraestrutura.")
		}

		return utils.JSONError(c, fiber.StatusInternalServerError, "Erro ao processar o disparo: "+err.Error())
	}

	return utils.JSONSuccess(c, fiber.StatusOK, fiber.Map{
		"message": "Broadcast concluído com sucesso",
		"metrics": metrics,
	})
}

// HandleGetBroadcastPageData maps GET /api/ui/broadcast-page/:guildId
func (h *BroadcastHandler) HandleGetBroadcastPageData(c *fiber.Ctx) error {
	guildID := c.Params("id")
	if guildID == "" {
		return utils.JSONError(c, fiber.StatusBadRequest, "O parâmetro guildId é obrigatório")
	}

	data, err := h.broadcastUsecase.GetBroadcastPageData(guildID)
	if err != nil {
		log.Printf("❌ ERRO [HandleGetBroadcastPageData] guilda %s: %v", guildID, err)
		return utils.JSONError(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.JSONSuccess(c, fiber.StatusOK, data)
}

// HandleCreateBroadcast maps POST /api/broadcasts
func (h *BroadcastHandler) HandleCreateBroadcast(c *fiber.Ctx) error {
	var record pocketbase.BroadcastRecord
	if err := c.BodyParser(&record); err != nil {
		return utils.JSONError(c, fiber.StatusBadRequest, "Corpo da requisição JSON inválido ou malformado")
	}

	if record.Content == "" {
		return utils.JSONError(c, fiber.StatusBadRequest, "O conteúdo do comunicado não pode ser vazio")
	}

	if record.TargetType != "public" && record.TargetType != "private" {
		return utils.JSONError(c, fiber.StatusBadRequest, "O tipo de destino deve ser 'public' ou 'private'")
	}

	if record.ScheduleTime == "" {
		return utils.JSONError(c, fiber.StatusBadRequest, "A data/hora do disparo (schedule_time) é obrigatória")
	}

	if record.GuildID == "" {
		return utils.JSONError(c, fiber.StatusBadRequest, "O identificador da guilda (guild_id) é obrigatório")
	}

	err := h.broadcastUsecase.CreateBroadcast(&record)
	if err != nil {
		log.Printf("❌ ERRO [HandleCreateBroadcast] salvando agendamento: %v", err)
		return utils.JSONError(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.JSONSuccess(c, fiber.StatusCreated, record)
}

// HandleCancelBroadcast maps DELETE /api/broadcasts/:id
func (h *BroadcastHandler) HandleCancelBroadcast(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return utils.JSONError(c, fiber.StatusBadRequest, "O parâmetro id da mensagem é obrigatório na rota")
	}

	err := h.broadcastUsecase.CancelBroadcast(id)
	if err != nil {
		log.Printf("❌ ERRO [HandleCancelBroadcast] cancelando msg %s: %v", id, err)
		return utils.JSONError(c, fiber.StatusBadRequest, err.Error())
	}

	return utils.JSONError(c, fiber.StatusOK, "Agendamento de comunicado cancelado e excluído com sucesso")
}

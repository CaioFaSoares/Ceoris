package utils

import "github.com/gofiber/fiber/v2"

// JSONSuccess retorna uma resposta envelopada em { "data": ... }
func JSONSuccess(c *fiber.Ctx, code int, data interface{}) error {
	return c.Status(code).JSON(fiber.Map{
		"data": data,
	})
}

// JSONError retorna uma resposta de erro envelopada em { "error": ... }
func JSONError(c *fiber.Ctx, code int, message string) error {
	return c.Status(code).JSON(fiber.Map{
		"error": message,
	})
}

package middleware

import (
	"os"

	"github.com/gofiber/fiber/v2"
)

func HasAuthorization(c *fiber.Ctx) error {
	apiKey := os.Getenv("API_KEY")
	if apiKey != c.Get("x-api-key") {
		return c.Status(403).JSON(fiber.Map{
			"message": "You don't have permission to do that action",
		})
	}

	return c.Next()
}

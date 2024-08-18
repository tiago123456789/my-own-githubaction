package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/gofiber/fiber/v2"
)

func HasValidSecret(c *fiber.Ctx) error {
	signature := c.Get("X-Hub-Signature-256")
	secret := c.Params("hash")

	if len(secret) == 0 {
		return c.Status(403).JSON(fiber.Map{
			"message": "You don't have permission to do that action",
		})
	}

	if len(signature) == 0 {
		return c.Status(403).JSON(fiber.Map{
			"message": "You don't have permission to do that action",
		})
	}

	body := c.Body()
	reader := bytes.NewReader(body)
	body, err := ioutil.ReadAll(reader)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"message": fmt.Sprintf("could not read request body: %v", err),
		})
	}

	io.NopCloser(bytes.NewBuffer(body))
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expectedMAC := mac.Sum(nil)
	expectedSignature := "sha256=" + hex.EncodeToString(expectedMAC)
	if !hmac.Equal([]byte(expectedSignature), []byte(signature)) {
		return c.Status(403).JSON(fiber.Map{
			"message": "You don't have permission to do that action",
		})
	}

	return c.Next()
}

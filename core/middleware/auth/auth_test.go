package auth

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func TestAuthMiddleware(t *testing.T) {
	app := fiber.New()
	
	cfg := Config{ApiKey: "secret"}
	app.Use(New(cfg))
	
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello")
	})

	// Case 1: No Key
	req := httptest.NewRequest("GET", "/", nil)
	resp, _ := app.Test(req)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

	// Case 2: Wrong Key
	req = httptest.NewRequest("GET", "/", nil)
	req.Header.Set(HeaderKey, "wrong")
	resp, _ = app.Test(req)
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)

	// Case 3: Correct Key
	req = httptest.NewRequest("GET", "/", nil)
	req.Header.Set(HeaderKey, "secret")
	resp, _ = app.Test(req)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

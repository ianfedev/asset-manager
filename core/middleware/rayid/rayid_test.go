package rayid

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func TestRayIDMiddleware(t *testing.T) {
	app := fiber.New()
	
	app.Use(New())
	
	app.Get("/", func(c *fiber.Ctx) error {
		rid := c.Locals(ContextKey)
		assert.NotNil(t, rid)
		assert.NotEmpty(t, rid)
		return c.SendString("ok")
	})

	req := httptest.NewRequest("GET", "/", nil)
	resp, _ := app.Test(req)
	
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
	assert.NotEmpty(t, resp.Header.Get(HeaderKey))
}

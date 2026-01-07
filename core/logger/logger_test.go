package logger

import (
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNew(t *testing.T) {
	t.Run("Console Debug", func(t *testing.T) {
		cfg := &Config{
			Level:  "debug",
			Format: "console",
		}
		l, err := New(cfg)
		assert.NoError(t, err)
		assert.NotNil(t, l)
	})

	t.Run("JSON Info", func(t *testing.T) {
		cfg := &Config{
			Level:  "info",
			Format: "json",
		}
		l, err := New(cfg)
		assert.NoError(t, err)
		assert.NotNil(t, l)
	})
}

func TestWithRayID(t *testing.T) {
	app := fiber.New()
	baseLogger := zap.NewExample()

	t.Run("With RayId", func(t *testing.T) {
		app.Get("/with", func(c *fiber.Ctx) error {
			c.Locals("ray_id", "test-id")
			l := WithRayID(baseLogger, c)
			assert.NotNil(t, l)
			assert.NotEqual(t, baseLogger, l) // Should be wrapped (clone)
			return c.SendStatus(200)
		})

		req := httptest.NewRequest("GET", "/with", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	})

	t.Run("Without RayId", func(t *testing.T) {
		app.Get("/without", func(c *fiber.Ctx) error {
			// No ray_id set
			l := WithRayID(baseLogger, c)
			assert.NotNil(t, l)
			assert.Equal(t, baseLogger, l) // Should return same logger if empty
			return c.SendStatus(200)
		})

		req := httptest.NewRequest("GET", "/without", nil)
		resp, err := app.Test(req)
		assert.NoError(t, err)
		assert.Equal(t, 200, resp.StatusCode)
	})
}

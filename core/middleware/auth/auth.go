package auth

import (
	"github.com/gofiber/fiber/v2"
)

const (
	// HeaderKey is the header key for API Key.
	HeaderKey = "X-API-Key"
)

// Config defines the config for Auth middleware.
type Config struct {
	// ApiKey is the secret key required to access the API.
	ApiKey string
}

// New creates a new Auth middleware.
// It checks the X-API-Key header against the configured key.
func New(cfg Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Skip if no key is configured (dev mode or strictly public, but better to enforce)
		// However, requirement says "protect every request", so if key is empty we might fail or allow.
		// Assuming enforce: if config is empty, maybe block all? Or log warning? 
		// For now, simple check.
		
		key := c.Get(HeaderKey)
		
		if cfg.ApiKey == "" {
			// Fail safe: if server has no key configured, deny everything to prevent accidental exposure?
			// Or allow? Given requirements "protect", likely fail safe.
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Server configuration error: No API Key configured",
			})
		}

		if key != cfg.ApiKey {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Unauthorized: Invalid or missing API Key",
			})
		}

		return c.Next()
	}
}

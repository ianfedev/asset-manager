package rayid

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

const (
	// HeaderKey is the header key for Ray ID.
	HeaderKey = "X-Ray-ID"
	// ContextKey is the key used to store Ray ID in Fiber locals.
	ContextKey = "ray_id"
)

// New creates a new Ray ID middleware.
// It generates a unique ID for each request and sets it in the response header and context.
func New() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Generate UUID
		rid := uuid.New().String()

		// Set for internal use (Logger, etc.)
		c.Locals(ContextKey, rid)

		// Set Header
		c.Set(HeaderKey, rid)

		return c.Next()
	}
}

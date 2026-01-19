package furniture

import (
	"asset-manager/core/storage/mocks"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func setupTestApp(h *Handler) (*fiber.App, *mocks.Client, *zap.Logger) {
	app := fiber.New()
	h.RegisterRoutes(app)
	return app, nil, nil
}

func TestHandler_HandleGetFurnitureDetail(t *testing.T) {
	mockClient := new(mocks.Client)
	logger := zap.NewNop()
	db, _ := setupMockDB(t)
	svc := NewService(mockClient, "test-bucket", logger, db, "arcturus")
	handler := NewHandler(svc)

	// Mock BucketExists which is called early in integrity check
	mockClient.On("BucketExists", mock.Anything, "test-bucket").Return(false, assert.AnError)

	app, _, _ := setupTestApp(handler)

	req := httptest.NewRequest("GET", "/furniture/test_item", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	// It should fail because BucketExists fails, returning 500
	assert.Equal(t, 500, resp.StatusCode)
}

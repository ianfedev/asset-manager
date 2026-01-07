package loader

import (
	"errors"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockFeatureLoader ...
type MockFeatureLoader struct {
	mock.Mock
}

func (m *MockFeatureLoader) Name() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockFeatureLoader) IsEnabled() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockFeatureLoader) Load(app fiber.Router) error {
	args := m.Called(app)
	return args.Error(0)
}

func TestManager(t *testing.T) {
	app := fiber.New()

	t.Run("LoadAll Success", func(t *testing.T) {
		manager := NewManager()
		loader1 := new(MockFeatureLoader)
		loader2 := new(MockFeatureLoader)

		// Loader 1: Enabled
		loader1.On("IsEnabled").Return(true)
		loader1.On("Name").Return("feature1")
		loader1.On("Load", app).Return(nil)

		// Loader 2: Disabled
		loader2.On("IsEnabled").Return(false)
		loader2.On("Name").Return("feature2")
		// Load should NOT be called

		manager.Register(loader1)
		manager.Register(loader2)

		err := manager.LoadAll(app)
		assert.NoError(t, err)

		loader1.AssertExpectations(t)
		loader2.AssertExpectations(t)
	})

	t.Run("LoadAll Error", func(t *testing.T) {
		manager := NewManager()
		loader := new(MockFeatureLoader)

		loader.On("IsEnabled").Return(true)
		loader.On("Name").Return("feature_fail")
		loader.On("Load", app).Return(errors.New("fail"))

		manager.Register(loader)

		err := manager.LoadAll(app)
		assert.Error(t, err)
		assert.Equal(t, "fail", err.Error())
	})
}

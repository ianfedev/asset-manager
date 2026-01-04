package loader

import (
	"log"

	"github.com/gofiber/fiber/v2"
)

// FeatureLoader interface that every feature must implement.
type FeatureLoader interface {
	Name() string
	IsEnabled() bool
	Load(app *fiber.App) error
}

// Manager handles the registration and loading of features.
type Manager struct {
	loaders []FeatureLoader
}

// NewManager creates a new loader manager.
func NewManager() *Manager {
	return &Manager{
		loaders: make([]FeatureLoader, 0),
	}
}

// Register adds a loader to the manager.
func (m *Manager) Register(l FeatureLoader) {
	m.loaders = append(m.loaders, l)
}

// LoadAll iterates through all registered loaders and loads them if enabled.
func (m *Manager) LoadAll(app *fiber.App) error {
	for _, l := range m.loaders {
		if l.IsEnabled() {
			log.Printf("Loading feature: %s", l.Name())
			if err := l.Load(app); err != nil {
				return err
			}
		} else {
			log.Printf("Skipping disabled feature: %s", l.Name())
		}
	}
	return nil
}

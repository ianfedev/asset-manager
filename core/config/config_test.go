package config

import (
	"reflect"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

// TestConfigDefaults ensures that every field in the Config struct has a default value set.
func TestConfigDefaults(t *testing.T) {
	// Initialize Viper
	v := viper.New()
	
	// Bind defaults using our helper function
	bindValues(v, Config{}, "")

	// Validate defaults
	validateDefaults(t, v, reflect.TypeOf(Config{}), "")
}

func validateDefaults(t *testing.T, v *viper.Viper, typ reflect.Type, prefix string) {
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		tag := field.Tag.Get("mapstructure")
		
		if tag == "" {
			continue
		}

		key := tag
		if prefix != "" {
			key = prefix + "." + tag
		}
		
		// If nested struct, recurse
		if field.Type.Kind() == reflect.Struct {
			validateDefaults(t, v, field.Type, key)
			continue
		}

		// Check if the default value is set in Viper, unless the default tag is explicitly empty
		defaultValue := field.Tag.Get("default")
		if defaultValue == "" {
			continue
		}
		
		if !v.IsSet(key) {
			t.Errorf("Default value not set for key: %s (Field: %s)", key, field.Name)
		}
	}
}

// TestLoadConfig checks if config loads correctly.
func TestLoadConfig(t *testing.T) {
	// Create a temporary .env file or just test defaults
	config, err := LoadConfig(".")
	assert.NoError(t, err)
	assert.NotNil(t, config)
	
	// Check defaults
	assert.Equal(t, "8080", config.Server.Port)
	assert.Equal(t, "minioadmin", config.Storage.AccessKey)
	assert.Equal(t, "info", config.Log.Level)
	assert.Equal(t, "json", config.Log.Format)
}

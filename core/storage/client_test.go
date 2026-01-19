package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{
			name: "Valid Config HTTP",
			cfg: Config{
				Endpoint:  "localhost:9000",
				AccessKey: "minio",
				SecretKey: "minio123",
				UseSSL:    false,
			},
			wantErr: false,
		},
		{
			name: "Valid Config HTTPS",
			cfg: Config{
				Endpoint:  "https://s3.amazonaws.com",
				AccessKey: "key",
				SecretKey: "secret",
				UseSSL:    true,
			},
			wantErr: false,
		},
		{
			name: "Empty Endpoint",
			cfg: Config{
				Endpoint: "", // Minio might error or default, usually errors on empty host
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.cfg)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, client)
			}
		})
	}
}

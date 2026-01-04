package server

// Config holds configuration for the HTTP server.
type Config struct {
	// Port is the port where the server will listen.
	Port string `mapstructure:"port" default:"8080"`
	// ApiKey is the secret key required to access the API.
	ApiKey string `mapstructure:"api_key" default:""`
}

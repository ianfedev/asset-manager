package logger

// Config holds configuration for logging.
type Config struct {
	// Level defines the logging level (debug, info, warn, error).
	Level string `mapstructure:"level" default:"info"`
	// Format defines the logging output format (json, console).
	// Default is "json". Use "console" for pretty printing during development.
	Format string `mapstructure:"format" default:"json"`
}

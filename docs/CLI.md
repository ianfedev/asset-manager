# CLI Documentation

The Asset Manager uses a Command Line Interface (CLI) powered by Cobra.

## Commands

### `asset-manager` (Root)
Running the binary without arguments or with `--help` will display the help message.

### `asset-manager start`
Starts the HTTP server.
- Loads configuration from `.env` or environment variables.
- Initializes the Zap logger.
- Sets up the Fiber web framework.
- loads all enabled features via the loader system.

## Usage

```bash
# Display help
go run main.go --help

# Start the server
go run main.go start
```

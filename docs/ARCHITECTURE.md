# Architecture & Design Guidelines

## Folder Structure

The project is organized into two primary top-level directories to separate core infrastructure from business logic features.

### `core/`
The `core/` folder contains foundational code that is agnostic to specific business features. This includes:
- **Project Structure**: Shared utilities, configuration management, and helper functions.
- **Interfaces**: Definitions for external dependencies (e.g., FileSystem, S3Client, Logger) to ensure testability.
- **Middleware**: Common HTTP middleware (authentication, logging, recovery).

### `feature/`
The `feature/` folder contains domain-specific logic. Each feature should be self-contained in its own package.
- **Organization**: Grouped by functionality (e.g., `feature/assets`, `feature/upload`).
- **Components**: Handlers, services, and models specific to that feature.

## Coding Standards

### Code Style
- **Self-Explanatory**: Code must be clear and readable without relying on comments inside functions to explain *what* is happening. Comments should explain *why* something is done if it's non-obvious.
- **No Unnecessary Comments**: Avoid "noise" comments like `// Loop through items` or `// Return result`.

### Documentation (Godocs)
All exported functions, types, constants, variables, **and struct fields** must have Godoc comments.
- Comments should be complete sentences starting with the name of the symbol.
- Package-level documentation should exist in a `doc.go` file or the primary source file of the package.

Example:
```go
// AssetService handles the retrieval and storage of asset files.
type AssetService struct {
    // TimeoutSeconds defines the maximum duration for a request.
    TimeoutSeconds int
}

// GetAsset retrieves an asset by its ID.
// It returns an error if the asset is not found.
func (s *AssetService) GetAsset(id string) (*Asset, error) {
    // ...
}
```

### Configuration
- **Structure**: Configuration must be partialized using nested structs (e.g., `StorageConfig`, `ServerConfig`) to group related settings.
- **Defaults**: Default values must be defined using a `default` struct tag.
- **Naming**: Use generic names for technologies where possible (e.g., use `Storage` instead of `Minio`).

### Testing Strategy

The architecture prioritizes testability through dependency injection and interface abstraction.

#### Unit Testing
- **Requirement**: All business logic must be unit testable.
- **Mocking**: Dependencies should be defined as interfaces in the `core` or `feature` packages.
- **Coverage**: Aim for high code coverage on logic-heavy components.
- **Config Testing**: Configuration defaults must be dynamically verified in unit tests.

#### Integration Testing
- **Requirement**: The system must support integration testing.
- **Scope**: Meaningful flows (e.g., "Upload file -> Check S3 -> Verify Metadata").

## Development Workflow
1. Define the feature package in `feature/`.
2. Strict adherence to defining interfaces for any external interaction.
3. Write unit tests alongside the implementation.
4. Ensure Godocs are present for all public symbols.

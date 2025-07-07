# Architecture Guidelines

## Clean Architecture Layers

### 1. Domain Layer (`internal/domain/`)
- **Purpose**: Core business logic and rules
- **Contents**: Entities, Value Objects, Repository Interfaces
- **Dependencies**: None (pure Go)
- **Example**: Transaction entity with business rules

### 2. Application Layer (`internal/application/`)
- **Purpose**: Use cases and application-specific business rules
- **Contents**: Use Cases, Application Services
- **Dependencies**: Domain layer only
- **Example**: CreateTransaction use case

### 3. Infrastructure Layer (`internal/infrastructure/`)
- **Purpose**: External services and frameworks
- **Contents**: Database implementations, External APIs
- **Dependencies**: All layers
- **Example**: MongoDB repository implementations

### 4. Interfaces Layer (`internal/interfaces/`)
- **Purpose**: User interface and interaction
- **Contents**: TUI, Input/Output handling
- **Dependencies**: Application layer
- **Example**: Account management screens

## Key Implementation Rules
1. **One entity per file**
2. **One use case per file**
3. **Functions under 50 lines**
4. **Clear, single responsibility**
5. **Proper error handling**

## Error Handling Strategy
```go
// Custom error types
type DomainError struct {
    Code    string
    Message string
}

type ValidationError struct {
    Field   string
    Message string
}

// Graceful error handling
if err != nil {
    return fmt.Errorf("failed to create account: %w", err)
}
```

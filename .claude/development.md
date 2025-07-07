# Development Workflow

## Setup

### Prerequisites
- Go 1.21+
- Docker & Docker Compose
- MongoDB (via Docker)

### Initial Setup
```bash
# Clone and setup
git clone <repo-url>
cd budget-cli

# Create .env file
cp .env.example .env

# Start MongoDB
docker-compose up -d

# Build application
go build -o financli cmd/main.go

# Run application
./financli
```

## Common Commands

### Development
```bash
# Run application
go run cmd/main.go

# Build for different architectures
go build -o bin/financli-linux cmd/main.go
go build -o bin/financli-mac cmd/main.go
go build -o bin/financli-windows.exe cmd/main.go

# Run tests
go test ./...
go test -v ./internal/domain/...

# Format code
go fmt ./...

# Lint code
golangci-lint run
```

### Database
```bash
# Start MongoDB
docker-compose up -d

# Connect to MongoDB
docker exec -it financli-mongodb mongosh -u admin -p password

# View logs
docker-compose logs -f mongodb

# Reset database
docker-compose down -v
docker-compose up -d
```

### Build & Deploy
```bash
# Build for production
go build -ldflags "-s -w" -o financli cmd/main.go

# Create release
make release

# Build Docker image
docker build -t budget-cli .
```

## Git Workflow

### Branch Strategy
- `main`: Production ready code
- `develop`: Integration branch
- `feature/*`: Feature branches
- `fix/*`: Bug fix branches

### Commit Messages
```bash
# Format: type(scope): description
feat(accounts): add account creation functionality
fix(transactions): resolve date parsing issue
docs(readme): update installation instructions
refactor(domain): improve entity structure
```

## Testing Strategy

### Unit Tests
```bash
# Run specific package tests
go test ./internal/domain/entity/

# Run with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Integration Tests
```bash
# Run integration tests (requires MongoDB)
go test -tags=integration ./...
```

## Code Quality

### Pre-commit Checklist
- [ ] Code formatted (`go fmt`)
- [ ] Tests passing (`go test`)
- [ ] Linter clean (`golangci-lint run`)
- [ ] Documentation updated
- [ ] Commit message follows convention

### Performance Considerations
- Use connection pooling for MongoDB
- Implement pagination for large datasets
- Add caching for frequently accessed data
- Profile memory usage for large operations

# Claude Code Documentation

This directory contains organized documentation for the Budget CLI project, designed to help Claude Code understand and work with the codebase effectively.

## Documentation Structure

- **`project-overview.md`**: High-level project information, tech stack, and structure
- **`architecture.md`**: Clean architecture guidelines and implementation rules
- **`database.md`**: MongoDB setup, schema design, and connection management
- **`development.md`**: Development workflow, commands, and best practices
- **`features.md`**: Current features, roadmap, and specifications

## Quick Reference

### Start Development
```bash
docker-compose up -d  # Start MongoDB
go run cmd/main.go    # Run application
```

### Build & Test
```bash
go build -o financli cmd/main.go
go test ./...
```

### Database Connection
```bash
# Ensure .env file exists with:
MONGODB_URI=mongodb://admin:password@localhost:27017/financli?authSource=admin
MONGODB_DATABASE=financli
```

---

*This documentation is maintained to provide context and guidance for development work on the Budget CLI project.*

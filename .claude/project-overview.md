# Budget CLI - Project Overview

## Current Stack
- **Language**: Go
- **Database**: MongoDB
- **CLI Framework**: Built with Go's standard library
- **Architecture**: Clean Architecture
- **Database Driver**: MongoDB Go Driver

## Project Structure
```
budget-cli/
├── cmd/
│   ├── main.go
│   └── demo/
├── internal/
│   ├── domain/
│   │   ├── entity/
│   │   ├── repository/
│   │   └── valueobject/
│   ├── application/
│   │   └── usecase/
│   ├── infrastructure/
│   │   ├── config/
│   │   └── persistence/
│   └── interfaces/
│       └── tui/
├── pkg/
└── scripts/
```

## Core Entities
- **Account**: Bank accounts with balance tracking
- **Transaction**: Financial transactions with categorization
- **Person**: People for shared transactions
- **Bill**: Recurring bills and expenses
- **Credit Card**: Credit card management

## Development Commands
- `go build`: Build the application
- `go run cmd/main.go`: Run the CLI
- `go test ./...`: Run tests
- `docker-compose up`: Start MongoDB

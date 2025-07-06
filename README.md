# FinanCLI - Personal Finance Manager

A comprehensive command-line personal finance management application built with Go, using Clean Architecture principles, Bubble Tea for the TUI, and MongoDB for data persistence.

## Features

- **Account Management**: Manage checking, savings, and investment accounts
- **Credit Card Management**: Track credit cards linked to accounts
- **Bill Management**: Organize bills with lifecycle tracking (open/closed/paid/overdue)
- **Transaction Management**: Record and categorize transactions with expense sharing
- **Person Management**: Manage people for expense sharing
- **Reports**: Generate detailed expense reports by bill or person
- **Dashboard**: Overview with ASCII charts and financial summaries

## Architecture

The project follows Clean Architecture principles:

```
financli/
├── cmd/                    # Application entry points
├── internal/
│   ├── domain/            # Business logic and entities
│   │   ├── entity/        # Domain entities
│   │   ├── valueobject/   # Value objects
│   │   └── repository/    # Repository interfaces
│   ├── application/       # Use cases
│   │   ├── usecase/       # Business use cases
│   │   └── service/       # Application services
│   ├── infrastructure/    # External dependencies
│   │   ├── persistence/   # Database implementations
│   │   └── config/        # Configuration
│   └── interfaces/        # User interfaces
│       └── tui/          # Terminal UI
└── pkg/                   # Shared packages
```

## Prerequisites

- Go 1.21+
- MongoDB
- Terminal with UTF-8 support

## Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd financli
```

2. Install dependencies:
```bash
go mod download
```

3. Set up environment variables:
```bash
export MONGODB_URI="mongodb://localhost:27017"
export MONGODB_DATABASE="financli"
```

## Usage

Run the application:
```bash
go run cmd/main.go
```

### Navigation

- **Number Keys (1-7)**: Switch between screens
- **Arrow Keys**: Navigate within screens
- **Enter**: Confirm actions
- **Esc**: Cancel operations
- **q/Ctrl+C**: Quit application

### Screens

1. **Dashboard**: Financial overview with charts
2. **Accounts**: Manage bank accounts
3. **Credit Cards**: Track credit card usage
4. **Bills**: Organize and pay bills
5. **Transactions**: Record expenses and income
6. **People**: Manage expense sharing contacts
7. **Reports**: View detailed financial reports

## Key Features

### Expense Sharing
- Split transactions with registered people
- Support for percentage-based or equal splits
- Automatic calculation of shared amounts

### Bill Management
- Track bill lifecycle (open → paid/overdue → closed)
- Automatic transaction assignment based on dates
- Payment tracking and status updates

### Financial Reports
- Shared expense reports by person
- Bill payment summaries
- Category-wise expense breakdowns

## Development

### Running Tests
```bash
go test ./...
```

### Building
```bash
go build -o financli cmd/main.go
```

## Technologies Used

- **Go 1.21+**: Programming language
- **Bubble Tea**: Terminal UI framework
- **Lip Gloss**: Terminal styling
- **MongoDB**: Data persistence
- **Clean Architecture**: Software design pattern
- **Dependency Injection**: For loose coupling
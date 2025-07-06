package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"financli/internal/application/usecase"
	"financli/internal/infrastructure/config"
	"financli/internal/infrastructure/persistence/mongodb"
	"financli/internal/interfaces/tui"
	
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	ctx := context.Background()
	
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}
	
	db, err := mongodb.NewConnection(mongodb.Config{
		URI:      cfg.MongoDB.URI,
		Database: cfg.MongoDB.Database,
	})
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}
	
	// Initialize repositories
	accountRepo := mongodb.NewAccountRepository(db)
	creditCardRepo := mongodb.NewCreditCardRepository(db)
	personRepo := mongodb.NewPersonRepository(db)
	billRepo := mongodb.NewBillRepository(db)
	transactionRepo := mongodb.NewTransactionRepository(db)
	
	// Initialize use cases
	useCases := tui.UseCases{
		Account:     usecase.NewAccountUseCase(accountRepo),
		CreditCard:  usecase.NewCreditCardUseCase(creditCardRepo, accountRepo),
		Bill:        usecase.NewBillUseCase(billRepo),
		Transaction: usecase.NewTransactionUseCase(transactionRepo, accountRepo, creditCardRepo, billRepo),
		Person:      usecase.NewPersonUseCase(personRepo),
		Report:      usecase.NewReportUseCase(transactionRepo, personRepo, billRepo),
	}
	
	// Initialize and run TUI
	app := tui.NewApp(ctx, useCases)
	p := tea.NewProgram(app, tea.WithAltScreen())
	
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v", err)
		os.Exit(1)
	}
}
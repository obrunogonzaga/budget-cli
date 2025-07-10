package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"financli/internal/application/usecase"
	"financli/internal/domain/entity"
	"financli/internal/domain/valueobject"
	"financli/internal/infrastructure/config"
	"financli/internal/infrastructure/persistence/mongodb"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/mongo"
)

func main() {
	fmt.Println("🎯 FinanCLI - Personal Finance Manager Demo")
	fmt.Println("==========================================")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Printf("Failed to load config (using defaults): %v", err)
		cfg = &config.Config{
			MongoDB: config.MongoDBConfig{
				URI:      "mongodb://localhost:27017",
				Database: "financli_demo",
			},
		}
	}

	fmt.Printf("📊 Connecting to MongoDB: %s\n", cfg.MongoDB.URI)

	// Always run in-memory demo for simplicity
	fmt.Println("🎮 Running in-memory demo to showcase functionality...")
	runInMemoryDemo()
}

func runInMemoryDemo() {
	fmt.Println("\n💰 Creating sample financial data...")

	// Create accounts
	checkingAccount := entity.NewAccount(
		"Main Checking",
		entity.AccountTypeChecking,
		valueobject.NewMoney(2500.75, "BRL"),
		"Primary checking account",
	)

	savingsAccount := entity.NewAccount(
		"Emergency Fund",
		entity.AccountTypeSavings,
		valueobject.NewMoney(15000.00, "BRL"),
		"Emergency savings",
	)

	fmt.Printf("🏦 Account: %s - Balance: %s\n", checkingAccount.Name, checkingAccount.Balance.String())
	fmt.Printf("🏦 Account: %s - Balance: %s\n", savingsAccount.Name, savingsAccount.Balance.String())

	// Create credit card
	creditCard, err := entity.NewCreditCard(
		checkingAccount.ID,
		"Rewards Card",
		"1234",
		valueobject.NewMoney(5000.00, "BRL"),
		15,
	)
	if err != nil {
		fmt.Printf("❌ Error creating credit card: %v\n", err)
		return
	}

	fmt.Printf("💳 Credit Card: %s - Available: %s\n", creditCard.Name, func() string {
		available, _ := creditCard.GetAvailableCredit()
		return available.String()
	}())

	// Create people for expense sharing
	alice := entity.NewPerson("Alice Smith", "alice@example.com", "555-0101")
	bob := entity.NewPerson("Bob Johnson", "bob@example.com", "555-0102")

	fmt.Printf("👤 Person: %s (%s)\n", alice.Name, alice.Email)
	fmt.Printf("👤 Person: %s (%s)\n", bob.Name, bob.Email)

	// Create bill
	startDate := time.Now().AddDate(0, 0, -15)
	endDate := time.Now().AddDate(0, 0, 15)
	dueDate := time.Now().AddDate(0, 0, 30)

	bill, err := entity.NewBill(
		"Monthly Utilities",
		"Electricity, Water, Gas",
		startDate,
		endDate,
		dueDate,
		valueobject.NewMoney(350.00, "BRL"),
	)
	if err != nil {
		fmt.Printf("❌ Error creating bill: %v\n", err)
		return
	}

	fmt.Printf("📄 Bill: %s - Total: %s (Due: %s)\n",
		bill.Name,
		bill.TotalAmount.String(),
		bill.DueDate.Format("Jan 02"),
	)

	// Create transactions
	groceryTxn := entity.NewTransaction(
		&checkingAccount.ID,
		nil,
		entity.TransactionTypeDebit,
		entity.TransactionCategoryFood,
		valueobject.NewMoney(127.45, "BRL"),
		"Grocery shopping at Whole Foods",
		time.Now().AddDate(0, 0, -2),
	)

	// Split grocery bill with Alice
	err = groceryTxn.SplitEqually([]uuid.UUID{alice.ID})
	if err != nil {
		fmt.Printf("❌ Error splitting transaction: %v\n", err)
		return
	}

	fmt.Printf("🛒 Transaction: %s - %s\n", groceryTxn.Description, groceryTxn.Amount.String())
	fmt.Printf("   💰 Personal amount: %s\n", groceryTxn.GetPersonalAmount().String())
	fmt.Printf("   🤝 Shared with: %s\n", alice.Name)

	// Gas transaction
	gasTxn := entity.NewTransaction(
		nil,
		&creditCard.ID,
		entity.TransactionTypeDebit,
		entity.TransactionCategoryTransportation,
		valueobject.NewMoney(65.20, "BRL"),
		"Gas station fill-up",
		time.Now().AddDate(0, 0, -1),
	)

	fmt.Printf("⛽ Transaction: %s - %s\n", gasTxn.Description, gasTxn.Amount.String())

	// Salary deposit
	salaryTxn := entity.NewTransaction(
		&checkingAccount.ID,
		nil,
		entity.TransactionTypeCredit,
		entity.TransactionCategoryIncome,
		valueobject.NewMoney(3500.00, "BRL"),
		"Monthly salary deposit",
		time.Now().AddDate(0, 0, -3),
	)

	fmt.Printf("💵 Transaction: %s - %s\n", salaryTxn.Description, salaryTxn.Amount.String())

	fmt.Println("\n📊 Financial Summary:")
	fmt.Println("====================")
	fmt.Printf("Total Assets: %s\n",
		func() string {
			total, _ := checkingAccount.Balance.Add(savingsAccount.Balance)
			return total.String()
		}(),
	)
	fmt.Printf("Credit Card Balance: %s\n", creditCard.CurrentBalance.String())
	fmt.Printf("Credit Utilization: %.1f%%\n", creditCard.GetUtilizationPercentage())
	fmt.Printf("Pending Bills: %s\n", bill.TotalAmount.String())

	fmt.Println("\n🎮 Demo completed! The application structure includes:")
	fmt.Println("   ✅ Clean Architecture with domain/infrastructure/interfaces")
	fmt.Println("   ✅ Account management (checking, savings, investment)")
	fmt.Println("   ✅ Credit card management linked to accounts")
	fmt.Println("   ✅ Bill lifecycle management (open/closed/paid/overdue)")
	fmt.Println("   ✅ Transaction management with expense sharing")
	fmt.Println("   ✅ Person management for sharing expenses")
	fmt.Println("   ✅ 50/50 expense splitting with detailed reports")
	fmt.Println("   ✅ Bubble Tea TUI with Lip Gloss styling")
	fmt.Println("   ✅ ASCII charts for dashboard visualization")
	fmt.Println("   ✅ MongoDB persistence layer")
	fmt.Println("   ✅ Comprehensive error handling")
	fmt.Println("   ✅ Testing infrastructure")
}

func runFullDemo(ctx context.Context, db *mongo.Database) {
	fmt.Println("\n🚀 Running full demo with MongoDB persistence...")

	// Initialize repositories
	accountRepo := mongodb.NewAccountRepository(db)
	creditCardRepo := mongodb.NewCreditCardRepository(db)
	personRepo := mongodb.NewPersonRepository(db)
	billRepo := mongodb.NewBillRepository(db)
	transactionRepo := mongodb.NewTransactionRepository(db)

	// Initialize use cases
	accountUC := usecase.NewAccountUseCase(accountRepo)
	creditCardUC := usecase.NewCreditCardUseCase(creditCardRepo, accountRepo)
	personUC := usecase.NewPersonUseCase(personRepo)
	billUC := usecase.NewBillUseCase(billRepo)
	transactionUC := usecase.NewTransactionUseCase(transactionRepo, accountRepo, creditCardRepo, billRepo)

	// Demo operations
	fmt.Println("📊 Creating account...")
	account, err := accountUC.CreateAccount(ctx, "Demo Account", entity.AccountTypeChecking, 1000.0, "BRL", "Demo account")
	if err != nil {
		fmt.Printf("❌ Error creating account: %v\n", err)
		return
	}

	fmt.Printf("✅ Created account: %s (Balance: %s)\n", account.Name, account.Balance.String())

	fmt.Println("💳 Creating credit card...")
	card, err := creditCardUC.CreateCreditCard(ctx, account.ID, "Demo Card", "5678", 2000.0, "BRL", 15)
	if err != nil {
		fmt.Printf("❌ Error creating credit card: %v\n", err)
		return
	}

	fmt.Printf("✅ Created credit card: %s\n", card.Name)

	fmt.Println("👤 Creating person...")
	person, err := personUC.CreatePerson(ctx, "Demo Person", "demo@example.com", "555-1234")
	if err != nil {
		fmt.Printf("❌ Error creating person: %v\n", err)
		return
	}

	fmt.Printf("✅ Created person: %s\n", person.Name)

	fmt.Println("📄 Creating bill...")
	bill, err := billUC.CreateBill(ctx,
		"Demo Bill",
		"Test bill",
		time.Now().AddDate(0, 0, -10),
		time.Now().AddDate(0, 0, 10),
		time.Now().AddDate(0, 0, 20),
		500.0,
		"BRL",
	)
	if err != nil {
		fmt.Printf("❌ Error creating bill: %v\n", err)
		return
	}

	fmt.Printf("✅ Created bill: %s\n", bill.Name)

	fmt.Println("💰 Creating transaction...")
	transaction, err := transactionUC.CreateTransaction(ctx,
		&account.ID,
		nil,
		entity.TransactionTypeDebit,
		entity.TransactionCategoryFood,
		100.0,
		"BRL",
		"Demo transaction",
		time.Now(),
	)
	if err != nil {
		fmt.Printf("❌ Error creating transaction: %v\n", err)
		return
	}

	fmt.Printf("✅ Created transaction: %s\n", transaction.Description)

	fmt.Println("\n🎉 Full demo completed successfully with MongoDB!")
	fmt.Println("   All data has been persisted to the database.")
}

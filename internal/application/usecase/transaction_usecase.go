package usecase

import (
	"context"
	"fmt"
	"time"

	"financli/internal/domain/entity"
	"financli/internal/domain/repository"
	"financli/internal/domain/valueobject"
	"github.com/google/uuid"
)

type TransactionUseCase struct {
	transactionRepo       repository.TransactionRepository
	accountRepo           repository.AccountRepository
	creditCardRepo        repository.CreditCardRepository
	creditCardInvoiceRepo repository.CreditCardInvoiceRepository
	billRepo              repository.BillRepository
}

func NewTransactionUseCase(
	transactionRepo repository.TransactionRepository,
	accountRepo repository.AccountRepository,
	creditCardRepo repository.CreditCardRepository,
	billRepo repository.BillRepository,
) *TransactionUseCase {
	return &TransactionUseCase{
		transactionRepo: transactionRepo,
		accountRepo:     accountRepo,
		creditCardRepo:  creditCardRepo,
		billRepo:        billRepo,
	}
}

// NewTransactionUseCaseWithInvoice creates a new transaction use case with invoice support
func NewTransactionUseCaseWithInvoice(
	transactionRepo repository.TransactionRepository,
	accountRepo repository.AccountRepository,
	creditCardRepo repository.CreditCardRepository,
	creditCardInvoiceRepo repository.CreditCardInvoiceRepository,
	billRepo repository.BillRepository,
) *TransactionUseCase {
	return &TransactionUseCase{
		transactionRepo:       transactionRepo,
		accountRepo:           accountRepo,
		creditCardRepo:        creditCardRepo,
		creditCardInvoiceRepo: creditCardInvoiceRepo,
		billRepo:              billRepo,
	}
}

func (uc *TransactionUseCase) CreateTransaction(
	ctx context.Context,
	accountID *uuid.UUID,
	creditCardID *uuid.UUID,
	transactionType entity.TransactionType,
	category entity.TransactionCategory,
	amount float64,
	currency string,
	description string,
	date time.Time,
) (*entity.Transaction, error) {
	money := valueobject.NewMoney(amount, currency)
	transaction := entity.NewTransaction(accountID, creditCardID, transactionType, category, money, description, date)

	// Update account or credit card balance
	if accountID != nil {
		account, err := uc.accountRepo.FindByID(ctx, *accountID)
		if err != nil {
			return nil, fmt.Errorf("account not found: %w", err)
		}

		if transactionType == entity.TransactionTypeDebit {
			if err := account.Withdraw(money); err != nil {
				return nil, fmt.Errorf("failed to withdraw from account: %w", err)
			}
		} else {
			if err := account.Deposit(money); err != nil {
				return nil, fmt.Errorf("failed to deposit to account: %w", err)
			}
		}

		if err := uc.accountRepo.Update(ctx, account); err != nil {
			return nil, fmt.Errorf("failed to update account: %w", err)
		}
	}

	if creditCardID != nil {
		card, err := uc.creditCardRepo.FindByID(ctx, *creditCardID)
		if err != nil {
			return nil, fmt.Errorf("credit card not found: %w", err)
		}

		if transactionType == entity.TransactionTypeDebit {
			if err := card.Charge(money); err != nil {
				return nil, fmt.Errorf("failed to charge credit card: %w", err)
			}
		} else {
			if err := card.Payment(money); err != nil {
				return nil, fmt.Errorf("failed to apply payment to card: %w", err)
			}
		}

		if err := uc.creditCardRepo.Update(ctx, card); err != nil {
			return nil, fmt.Errorf("failed to update credit card: %w", err)
		}

		// Handle invoice assignment if invoice repository is available
		if uc.creditCardInvoiceRepo != nil {
			if err := uc.assignToInvoice(ctx, transaction, *creditCardID, transactionType == entity.TransactionTypeCredit); err != nil {
				// Log warning but don't fail the transaction
				fmt.Printf("Warning: failed to assign to invoice: %v\n", err)
			}
		}
	}

	// Auto-assign to bills if applicable
	if err := uc.autoAssignToBills(ctx, transaction); err != nil {
		// Log warning but don't fail the transaction
		fmt.Printf("Warning: failed to auto-assign to bills: %v\n", err)
	}

	if err := uc.transactionRepo.Create(ctx, transaction); err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	return transaction, nil
}

func (uc *TransactionUseCase) GetTransaction(ctx context.Context, id uuid.UUID) (*entity.Transaction, error) {
	return uc.transactionRepo.FindByID(ctx, id)
}

func (uc *TransactionUseCase) GetTransactionsByDateRange(ctx context.Context, startDate, endDate time.Time) ([]*entity.Transaction, error) {
	return uc.transactionRepo.FindByDateRange(ctx, startDate, endDate)
}

func (uc *TransactionUseCase) GetTransactionsByAccount(ctx context.Context, accountID uuid.UUID) ([]*entity.Transaction, error) {
	return uc.transactionRepo.FindByAccountID(ctx, accountID)
}

func (uc *TransactionUseCase) GetTransactionsByCreditCard(ctx context.Context, creditCardID uuid.UUID) ([]*entity.Transaction, error) {
	return uc.transactionRepo.FindByCreditCardID(ctx, creditCardID)
}

func (uc *TransactionUseCase) SplitTransactionEqually(ctx context.Context, transactionID uuid.UUID, personIDs []uuid.UUID) error {
	transaction, err := uc.transactionRepo.FindByID(ctx, transactionID)
	if err != nil {
		return err
	}

	if err := transaction.SplitEqually(personIDs); err != nil {
		return err
	}

	return uc.transactionRepo.Update(ctx, transaction)
}

func (uc *TransactionUseCase) AddSharedExpense(ctx context.Context, transactionID uuid.UUID, personID uuid.UUID, percentage float64) error {
	transaction, err := uc.transactionRepo.FindByID(ctx, transactionID)
	if err != nil {
		return err
	}

	if err := transaction.AddSharedExpense(personID, percentage); err != nil {
		return err
	}

	return uc.transactionRepo.Update(ctx, transaction)
}

func (uc *TransactionUseCase) autoAssignToBills(ctx context.Context, transaction *entity.Transaction) error {
	// Find bills that cover this transaction date
	bills, err := uc.billRepo.FindByDateRange(ctx, transaction.Date, transaction.Date)
	if err != nil {
		return err
	}

	// Find the most appropriate bill (e.g., shortest date range)
	var selectedBill *entity.Bill
	for _, bill := range bills {
		if bill.Status == entity.BillStatusOpen {
			if selectedBill == nil || bill.EndDate.Sub(bill.StartDate) < selectedBill.EndDate.Sub(selectedBill.StartDate) {
				selectedBill = bill
			}
		}
	}

	if selectedBill != nil {
		transaction.AssignToBill(selectedBill.ID)
	}

	return nil
}

func (uc *TransactionUseCase) assignToInvoice(ctx context.Context, transaction *entity.Transaction, creditCardID uuid.UUID, isPayment bool) error {
	// Find or create the current invoice for the transaction date
	var invoice *entity.CreditCardInvoice

	// First, try to find an invoice that contains this transaction date
	invoices, err := uc.creditCardInvoiceRepo.FindByCreditCard(ctx, creditCardID)
	if err != nil {
		return err
	}

	// Find the invoice that contains this transaction date
	for _, inv := range invoices {
		if inv.ContainsDate(transaction.Date) {
			invoice = inv
			break
		}
	}

	// If no invoice found, try to get the open invoice
	if invoice == nil {
		openInvoice, err := uc.creditCardInvoiceRepo.FindOpenInvoice(ctx, creditCardID)
		if err == nil && openInvoice != nil {
			invoice = openInvoice
		}
	}

	// If still no invoice, create one for the current month
	if invoice == nil {
		// Get credit card to access due day
		card, err := uc.creditCardRepo.FindByID(ctx, creditCardID)
		if err != nil {
			return fmt.Errorf("failed to get credit card for invoice creation: %w", err)
		}

		// Calculate invoice dates based on transaction date
		year, month := transaction.Date.Year(), transaction.Date.Month()
		referenceMonth := fmt.Sprintf("%04d-%02d", year, month)

		openingDate := time.Date(year, month, 1, 0, 0, 0, 0, transaction.Date.Location())
		closingDate := time.Date(year, month+1, 1, 0, 0, 0, 0, transaction.Date.Location()).AddDate(0, 0, -1)
		dueDate := time.Date(year, month+1, card.DueDay, 0, 0, 0, 0, transaction.Date.Location())

		// Get previous balance from the most recent closed invoice
		previousBalance := valueobject.NewMoney(0, card.CreditLimit.Currency())
		for _, inv := range invoices {
			if inv.IsClosed() && inv.ReferenceMonth < referenceMonth {
				previousBalance = inv.ClosingBalance
				break
			}
		}

		newInvoice, err := entity.NewCreditCardInvoice(creditCardID, referenceMonth, openingDate, closingDate, dueDate, previousBalance)
		if err != nil {
			return fmt.Errorf("failed to create invoice: %w", err)
		}

		if err := uc.creditCardInvoiceRepo.Create(ctx, newInvoice); err != nil {
			return fmt.Errorf("failed to save invoice: %w", err)
		}

		invoice = newInvoice
	}

	// Check if invoice is open
	if !invoice.IsOpen() {
		return fmt.Errorf("cannot add transaction to closed invoice")
	}

	// Add transaction to invoice
	if err := invoice.AddTransaction(transaction.ID, transaction.Amount, isPayment); err != nil {
		return err
	}

	// Update invoice
	if err := uc.creditCardInvoiceRepo.Update(ctx, invoice); err != nil {
		return fmt.Errorf("failed to update invoice: %w", err)
	}

	// Update transaction with invoice ID
	transaction.AssignToCreditCardInvoice(invoice.ID)

	return nil
}

// GetTransactionsByCreditCardInvoice returns all transactions for a specific invoice
func (uc *TransactionUseCase) GetTransactionsByCreditCardInvoice(ctx context.Context, invoiceID uuid.UUID) ([]*entity.Transaction, error) {
	return uc.transactionRepo.FindByCreditCardInvoiceID(ctx, invoiceID)
}

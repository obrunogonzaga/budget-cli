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

type CreditCardInvoiceUseCase struct {
	invoiceRepo    repository.CreditCardInvoiceRepository
	creditCardRepo repository.CreditCardRepository
}

func NewCreditCardInvoiceUseCase(invoiceRepo repository.CreditCardInvoiceRepository, creditCardRepo repository.CreditCardRepository) *CreditCardInvoiceUseCase {
	return &CreditCardInvoiceUseCase{
		invoiceRepo:    invoiceRepo,
		creditCardRepo: creditCardRepo,
	}
}

// CreateInvoice creates a new invoice for a credit card
func (uc *CreditCardInvoiceUseCase) CreateInvoice(ctx context.Context, creditCardID uuid.UUID, referenceMonth string, openingDate, closingDate, dueDate time.Time) (*entity.CreditCardInvoice, error) {
	// Verify credit card exists
	card, err := uc.creditCardRepo.FindByID(ctx, creditCardID)
	if err != nil {
		return nil, fmt.Errorf("credit card not found: %w", err)
	}

	// Check if invoice already exists for this month
	existing, _ := uc.invoiceRepo.FindByMonth(ctx, creditCardID, referenceMonth)
	if existing != nil {
		return nil, fmt.Errorf("invoice already exists for %s", referenceMonth)
	}

	// Get previous invoice to calculate previous balance
	previousBalance := valueobject.NewMoney(0, card.CreditLimit.Currency())

	// Find the most recent closed invoice
	invoices, err := uc.invoiceRepo.FindByCreditCard(ctx, creditCardID)
	if err == nil && len(invoices) > 0 {
		// Sort to find the most recent closed invoice
		for i := len(invoices) - 1; i >= 0; i-- {
			if invoices[i].IsClosed() && invoices[i].ReferenceMonth < referenceMonth {
				previousBalance = invoices[i].ClosingBalance
				break
			}
		}
	}

	invoice, err := entity.NewCreditCardInvoice(creditCardID, referenceMonth, openingDate, closingDate, dueDate, previousBalance)
	if err != nil {
		return nil, err
	}

	if err := uc.invoiceRepo.Create(ctx, invoice); err != nil {
		return nil, fmt.Errorf("failed to create invoice: %w", err)
	}

	return invoice, nil
}

// GetCurrentInvoice gets or creates the current open invoice for a credit card
func (uc *CreditCardInvoiceUseCase) GetCurrentInvoice(ctx context.Context, creditCardID uuid.UUID) (*entity.CreditCardInvoice, error) {
	// First try to find an open invoice
	invoice, err := uc.invoiceRepo.FindOpenInvoice(ctx, creditCardID)
	if err == nil && invoice != nil {
		return invoice, nil
	}

	// If no open invoice, create one for the current month
	card, err := uc.creditCardRepo.FindByID(ctx, creditCardID)
	if err != nil {
		return nil, fmt.Errorf("credit card not found: %w", err)
	}

	now := time.Now()
	year, month := now.Year(), now.Month()
	referenceMonth := fmt.Sprintf("%04d-%02d", year, month)

	// Calculate dates based on credit card due day
	openingDate := time.Date(year, month, 1, 0, 0, 0, 0, now.Location())
	closingDate := time.Date(year, month+1, 1, 0, 0, 0, 0, now.Location()).AddDate(0, 0, -1)

	// Due date is the card's due day of the next month
	dueDate := time.Date(year, month+1, card.DueDay, 0, 0, 0, 0, now.Location())

	return uc.CreateInvoice(ctx, creditCardID, referenceMonth, openingDate, closingDate, dueDate)
}

// CloseInvoice closes an invoice and optionally creates the next month's invoice
func (uc *CreditCardInvoiceUseCase) CloseInvoice(ctx context.Context, invoiceID uuid.UUID, createNext bool) error {
	invoice, err := uc.invoiceRepo.FindByID(ctx, invoiceID)
	if err != nil {
		return err
	}

	if err := invoice.Close(); err != nil {
		return err
	}

	if err := uc.invoiceRepo.Update(ctx, invoice); err != nil {
		return fmt.Errorf("failed to close invoice: %w", err)
	}

	// Create next month's invoice if requested
	if createNext {
		card, err := uc.creditCardRepo.FindByID(ctx, invoice.CreditCardID)
		if err != nil {
			return fmt.Errorf("credit card not found: %w", err)
		}

		// Parse current invoice month
		t, _ := time.Parse("2006-01", invoice.ReferenceMonth)
		nextMonth := t.AddDate(0, 1, 0)
		referenceMonth := nextMonth.Format("2006-01")

		year, month := nextMonth.Year(), nextMonth.Month()
		openingDate := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
		closingDate := time.Date(year, month+1, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, -1)
		dueDate := time.Date(year, month+1, card.DueDay, 0, 0, 0, 0, time.UTC)

		_, err = uc.CreateInvoice(ctx, invoice.CreditCardID, referenceMonth, openingDate, closingDate, dueDate)
		if err != nil {
			return fmt.Errorf("failed to create next invoice: %w", err)
		}
	}

	return nil
}

// ListInvoicesByCard lists all invoices for a credit card
func (uc *CreditCardInvoiceUseCase) ListInvoicesByCard(ctx context.Context, creditCardID uuid.UUID) ([]*entity.CreditCardInvoice, error) {
	return uc.invoiceRepo.FindByCreditCard(ctx, creditCardID)
}

// GetInvoiceByID gets a specific invoice
func (uc *CreditCardInvoiceUseCase) GetInvoiceByID(ctx context.Context, invoiceID uuid.UUID) (*entity.CreditCardInvoice, error) {
	return uc.invoiceRepo.FindByID(ctx, invoiceID)
}

// AddTransactionToInvoice adds a transaction to an invoice
func (uc *CreditCardInvoiceUseCase) AddTransactionToInvoice(ctx context.Context, invoiceID, transactionID uuid.UUID, amount float64, currency string, isPayment bool) error {
	invoice, err := uc.invoiceRepo.FindByID(ctx, invoiceID)
	if err != nil {
		return err
	}

	money := valueobject.NewMoney(amount, currency)
	if err := invoice.AddTransaction(transactionID, money, isPayment); err != nil {
		return err
	}

	return uc.invoiceRepo.Update(ctx, invoice)
}

// RemoveTransactionFromInvoice removes a transaction from an invoice
func (uc *CreditCardInvoiceUseCase) RemoveTransactionFromInvoice(ctx context.Context, invoiceID, transactionID uuid.UUID, amount float64, currency string, isPayment bool) error {
	invoice, err := uc.invoiceRepo.FindByID(ctx, invoiceID)
	if err != nil {
		return err
	}

	money := valueobject.NewMoney(amount, currency)
	if err := invoice.RemoveTransaction(transactionID, money, isPayment); err != nil {
		return err
	}

	return uc.invoiceRepo.Update(ctx, invoice)
}

// ProcessPayment processes a payment for an invoice
func (uc *CreditCardInvoiceUseCase) ProcessPayment(ctx context.Context, invoiceID uuid.UUID, amount float64, currency string) error {
	invoice, err := uc.invoiceRepo.FindByID(ctx, invoiceID)
	if err != nil {
		return err
	}

	// Add as a payment transaction
	transactionID := uuid.New() // This would normally come from the transaction creation
	money := valueobject.NewMoney(amount, currency)

	if err := invoice.AddTransaction(transactionID, money, true); err != nil {
		return err
	}

	// Check if invoice is fully paid
	if invoice.ClosingBalance.IsZero() || invoice.ClosingBalance.IsNegative() {
		if err := invoice.MarkAsPaid(); err != nil {
			return err
		}
	}

	return uc.invoiceRepo.Update(ctx, invoice)
}

// GetInvoicesByStatus gets all invoices with a specific status for a credit card
func (uc *CreditCardInvoiceUseCase) GetInvoicesByStatus(ctx context.Context, creditCardID uuid.UUID, status entity.InvoiceStatus) ([]*entity.CreditCardInvoice, error) {
	return uc.invoiceRepo.FindByStatus(ctx, creditCardID, status)
}

// UpdateOverdueInvoices checks and updates overdue invoices
func (uc *CreditCardInvoiceUseCase) UpdateOverdueInvoices(ctx context.Context, creditCardID uuid.UUID) error {
	// Get all closed invoices
	closedInvoices, err := uc.invoiceRepo.FindByStatus(ctx, creditCardID, entity.InvoiceStatusClosed)
	if err != nil {
		return err
	}

	now := time.Now()
	for _, invoice := range closedInvoices {
		if now.After(invoice.DueDate) && !invoice.ClosingBalance.IsZero() && !invoice.ClosingBalance.IsNegative() {
			invoice.Status = entity.InvoiceStatusOverdue
			if err := uc.invoiceRepo.Update(ctx, invoice); err != nil {
				return err
			}
		}
	}

	return nil
}

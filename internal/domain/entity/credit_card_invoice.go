package entity

import (
	"fmt"
	"time"

	"financli/internal/domain/valueobject"
	"github.com/google/uuid"
)

type InvoiceStatus string

const (
	InvoiceStatusOpen    InvoiceStatus = "open"
	InvoiceStatusClosed  InvoiceStatus = "closed"
	InvoiceStatusPaid    InvoiceStatus = "paid"
	InvoiceStatusOverdue InvoiceStatus = "overdue"
)

type CreditCardInvoice struct {
	ID              uuid.UUID
	CreditCardID    uuid.UUID
	ReferenceMonth  string // Format: "YYYY-MM"
	OpeningDate     time.Time
	ClosingDate     time.Time
	DueDate         time.Time
	PreviousBalance valueobject.Money
	TotalCharges    valueobject.Money
	TotalPayments   valueobject.Money
	ClosingBalance  valueobject.Money
	Status          InvoiceStatus
	TransactionIDs  []uuid.UUID
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func NewCreditCardInvoice(creditCardID uuid.UUID, referenceMonth string, openingDate, closingDate, dueDate time.Time, previousBalance valueobject.Money) (*CreditCardInvoice, error) {
	// Validate dates
	if closingDate.Before(openingDate) {
		return nil, fmt.Errorf("closing date cannot be before opening date")
	}
	
	if dueDate.Before(closingDate) {
		return nil, fmt.Errorf("due date cannot be before closing date")
	}
	
	// Validate reference month format
	if _, err := time.Parse("2006-01", referenceMonth); err != nil {
		return nil, fmt.Errorf("invalid reference month format, expected YYYY-MM")
	}
	
	now := time.Now()
	return &CreditCardInvoice{
		ID:              uuid.New(),
		CreditCardID:    creditCardID,
		ReferenceMonth:  referenceMonth,
		OpeningDate:     openingDate,
		ClosingDate:     closingDate,
		DueDate:         dueDate,
		PreviousBalance: previousBalance,
		TotalCharges:    valueobject.NewMoney(0, previousBalance.Currency()),
		TotalPayments:   valueobject.NewMoney(0, previousBalance.Currency()),
		ClosingBalance:  previousBalance,
		Status:          InvoiceStatusOpen,
		TransactionIDs:  []uuid.UUID{},
		CreatedAt:       now,
		UpdatedAt:       now,
	}, nil
}

func (i *CreditCardInvoice) AddTransaction(transactionID uuid.UUID, amount valueobject.Money, isPayment bool) error {
	if i.Status != InvoiceStatusOpen {
		return fmt.Errorf("cannot add transaction to %s invoice", i.Status)
	}
	
	// Add transaction ID to the list
	i.TransactionIDs = append(i.TransactionIDs, transactionID)
	
	// Update amounts
	if isPayment {
		newPayments, err := i.TotalPayments.Add(amount)
		if err != nil {
			return err
		}
		i.TotalPayments = newPayments
	} else {
		newCharges, err := i.TotalCharges.Add(amount)
		if err != nil {
			return err
		}
		i.TotalCharges = newCharges
	}
	
	// Recalculate closing balance
	if err := i.recalculateBalance(); err != nil {
		return err
	}
	
	i.UpdatedAt = time.Now()
	return nil
}

func (i *CreditCardInvoice) RemoveTransaction(transactionID uuid.UUID, amount valueobject.Money, isPayment bool) error {
	if i.Status != InvoiceStatusOpen {
		return fmt.Errorf("cannot remove transaction from %s invoice", i.Status)
	}
	
	// Remove transaction ID from the list
	newTransactionIDs := []uuid.UUID{}
	found := false
	for _, id := range i.TransactionIDs {
		if id != transactionID {
			newTransactionIDs = append(newTransactionIDs, id)
		} else {
			found = true
		}
	}
	
	if !found {
		return fmt.Errorf("transaction not found in invoice")
	}
	
	i.TransactionIDs = newTransactionIDs
	
	// Update amounts
	if isPayment {
		newPayments, err := i.TotalPayments.Subtract(amount)
		if err != nil {
			return err
		}
		i.TotalPayments = newPayments
	} else {
		newCharges, err := i.TotalCharges.Subtract(amount)
		if err != nil {
			return err
		}
		i.TotalCharges = newCharges
	}
	
	// Recalculate closing balance
	if err := i.recalculateBalance(); err != nil {
		return err
	}
	
	i.UpdatedAt = time.Now()
	return nil
}

func (i *CreditCardInvoice) recalculateBalance() error {
	// Closing Balance = Previous Balance + Total Charges - Total Payments
	balanceWithCharges, err := i.PreviousBalance.Add(i.TotalCharges)
	if err != nil {
		return err
	}
	
	closingBalance, err := balanceWithCharges.Subtract(i.TotalPayments)
	if err != nil {
		return err
	}
	
	i.ClosingBalance = closingBalance
	return nil
}

func (i *CreditCardInvoice) Close() error {
	if i.Status != InvoiceStatusOpen {
		return fmt.Errorf("invoice is already %s", i.Status)
	}
	
	i.Status = InvoiceStatusClosed
	i.UpdatedAt = time.Now()
	
	// Check if it should be marked as overdue
	i.updateStatusIfOverdue()
	
	return nil
}

func (i *CreditCardInvoice) MarkAsPaid() error {
	if i.Status == InvoiceStatusPaid {
		return fmt.Errorf("invoice is already paid")
	}
	
	if !i.ClosingBalance.IsZero() && !i.ClosingBalance.IsNegative() {
		return fmt.Errorf("invoice still has outstanding balance of %s", i.ClosingBalance.String())
	}
	
	i.Status = InvoiceStatusPaid
	i.UpdatedAt = time.Now()
	return nil
}

func (i *CreditCardInvoice) updateStatusIfOverdue() {
	if i.Status == InvoiceStatusClosed && time.Now().After(i.DueDate) && !i.ClosingBalance.IsZero() && !i.ClosingBalance.IsNegative() {
		i.Status = InvoiceStatusOverdue
	}
}

func (i *CreditCardInvoice) IsOpen() bool {
	return i.Status == InvoiceStatusOpen
}

func (i *CreditCardInvoice) IsClosed() bool {
	return i.Status == InvoiceStatusClosed || i.Status == InvoiceStatusPaid || i.Status == InvoiceStatusOverdue
}

func (i *CreditCardInvoice) GetStatementPeriod() string {
	return fmt.Sprintf("%s to %s", i.OpeningDate.Format("Jan 02"), i.ClosingDate.Format("Jan 02, 2006"))
}

func (i *CreditCardInvoice) GetDueDateFormatted() string {
	return i.DueDate.Format("January 02, 2006")
}

// Check if a transaction date falls within this invoice period
func (i *CreditCardInvoice) ContainsDate(date time.Time) bool {
	return (date.Equal(i.OpeningDate) || date.After(i.OpeningDate)) && 
	       (date.Equal(i.ClosingDate) || date.Before(i.ClosingDate))
}
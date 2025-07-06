package entity

import (
	"fmt"
	"time"

	"financli/internal/domain/valueobject"
	"github.com/google/uuid"
)

type BillStatus string

const (
	BillStatusOpen    BillStatus = "open"
	BillStatusClosed  BillStatus = "closed"
	BillStatusPaid    BillStatus = "paid"
	BillStatusOverdue BillStatus = "overdue"
)

type Bill struct {
	ID          uuid.UUID
	Name        string
	Description string
	StartDate   time.Time
	EndDate     time.Time
	DueDate     time.Time
	TotalAmount valueobject.Money
	PaidAmount  valueobject.Money
	Status      BillStatus
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func NewBill(name, description string, startDate, endDate, dueDate time.Time, totalAmount valueobject.Money) (*Bill, error) {
	if endDate.Before(startDate) {
		return nil, fmt.Errorf("end date cannot be before start date")
	}
	
	if dueDate.Before(endDate) {
		return nil, fmt.Errorf("due date cannot be before end date")
	}

	now := time.Now()
	return &Bill{
		ID:          uuid.New(),
		Name:        name,
		Description: description,
		StartDate:   startDate,
		EndDate:     endDate,
		DueDate:     dueDate,
		TotalAmount: totalAmount,
		PaidAmount:  valueobject.NewMoney(0, totalAmount.Currency()),
		Status:      BillStatusOpen,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

func (b *Bill) AddPayment(amount valueobject.Money) error {
	newPaidAmount, err := b.PaidAmount.Add(amount)
	if err != nil {
		return err
	}
	
	b.PaidAmount = newPaidAmount
	b.UpdatedAt = time.Now()
	
	b.updateStatus()
	return nil
}

func (b *Bill) updateStatus() {
	if b.PaidAmount.Equals(b.TotalAmount) {
		b.Status = BillStatusPaid
	} else if time.Now().After(b.DueDate) && !b.PaidAmount.Equals(b.TotalAmount) {
		b.Status = BillStatusOverdue
	}
}

func (b *Bill) Close() error {
	if b.Status == BillStatusPaid || b.Status == BillStatusClosed {
		return fmt.Errorf("bill is already %s", b.Status)
	}
	
	b.Status = BillStatusClosed
	b.UpdatedAt = time.Now()
	return nil
}

func (b *Bill) GetRemainingAmount() (valueobject.Money, error) {
	return b.TotalAmount.Subtract(b.PaidAmount)
}

func (b *Bill) IsFullyPaid() bool {
	return b.PaidAmount.Equals(b.TotalAmount)
}

func (b *Bill) GetPaymentPercentage() float64 {
	if b.TotalAmount.IsZero() {
		return 100
	}
	return (b.PaidAmount.Amount() / b.TotalAmount.Amount()) * 100
}
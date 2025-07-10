package entity

import (
	"fmt"
	"time"

	"financli/internal/domain/valueobject"
	"github.com/google/uuid"
)

type TransactionType string

const (
	TransactionTypeDebit  TransactionType = "debit"
	TransactionTypeCredit TransactionType = "credit"
)

type TransactionCategory string

const (
	TransactionCategoryFood           TransactionCategory = "food"
	TransactionCategoryTransportation TransactionCategory = "transportation"
	TransactionCategoryUtilities      TransactionCategory = "utilities"
	TransactionCategoryEntertainment  TransactionCategory = "entertainment"
	TransactionCategoryShopping       TransactionCategory = "shopping"
	TransactionCategoryHealthcare     TransactionCategory = "healthcare"
	TransactionCategoryEducation      TransactionCategory = "education"
	TransactionCategoryIncome         TransactionCategory = "income"
	TransactionCategoryTransfer       TransactionCategory = "transfer"
	TransactionCategoryOther          TransactionCategory = "other"
)

type Transaction struct {
	ID                  uuid.UUID
	AccountID           *uuid.UUID
	CreditCardID        *uuid.UUID
	CreditCardInvoiceID *uuid.UUID
	BillID              *uuid.UUID
	Type                TransactionType
	Category            TransactionCategory
	Amount              valueobject.Money
	Description         string
	Date                time.Time
	SharedWith          []SharedExpense
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type SharedExpense struct {
	PersonID   uuid.UUID
	Amount     valueobject.Money
	Percentage float64
}

func NewTransaction(
	accountID *uuid.UUID,
	creditCardID *uuid.UUID,
	transactionType TransactionType,
	category TransactionCategory,
	amount valueobject.Money,
	description string,
	date time.Time,
) *Transaction {
	now := time.Now()
	return &Transaction{
		ID:           uuid.New(),
		AccountID:    accountID,
		CreditCardID: creditCardID,
		Type:         transactionType,
		Category:     category,
		Amount:       amount,
		Description:  description,
		Date:         date,
		SharedWith:   []SharedExpense{},
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

func (t *Transaction) AssignToBill(billID uuid.UUID) {
	t.BillID = &billID
	t.UpdatedAt = time.Now()
}

func (t *Transaction) AssignToCreditCardInvoice(invoiceID uuid.UUID) {
	t.CreditCardInvoiceID = &invoiceID
	t.UpdatedAt = time.Now()
}

func (t *Transaction) AddSharedExpense(personID uuid.UUID, percentage float64) error {
	if percentage <= 0 || percentage > 100 {
		return fmt.Errorf("percentage must be between 0 and 100")
	}

	totalPercentage := percentage
	for _, shared := range t.SharedWith {
		totalPercentage += shared.Percentage
	}

	if totalPercentage > 100 {
		return fmt.Errorf("total shared percentage cannot exceed 100%%")
	}

	sharedAmount := t.Amount.Multiply(percentage / 100)

	t.SharedWith = append(t.SharedWith, SharedExpense{
		PersonID:   personID,
		Amount:     sharedAmount,
		Percentage: percentage,
	})

	t.UpdatedAt = time.Now()
	return nil
}

func (t *Transaction) SplitEqually(personIDs []uuid.UUID) error {
	if len(personIDs) == 0 {
		return fmt.Errorf("must provide at least one person to split with")
	}

	percentage := 50.0
	perPersonPercentage := percentage / float64(len(personIDs))

	t.SharedWith = []SharedExpense{}

	for _, personID := range personIDs {
		sharedAmount := t.Amount.Multiply(perPersonPercentage / 100)
		t.SharedWith = append(t.SharedWith, SharedExpense{
			PersonID:   personID,
			Amount:     sharedAmount,
			Percentage: perPersonPercentage,
		})
	}

	t.UpdatedAt = time.Now()
	return nil
}

func (t *Transaction) GetPersonalAmount() valueobject.Money {
	totalSharedPercentage := 0.0
	for _, shared := range t.SharedWith {
		totalSharedPercentage += shared.Percentage
	}

	personalPercentage := 100.0 - totalSharedPercentage
	return t.Amount.Multiply(personalPercentage / 100)
}

func (t *Transaction) ClearSharedExpenses() {
	t.SharedWith = []SharedExpense{}
	t.UpdatedAt = time.Now()
}

package entity

import (
	"fmt"
	"time"

	"financli/internal/domain/valueobject"
	"github.com/google/uuid"
)

type AccountType string

const (
	AccountTypeChecking   AccountType = "checking"
	AccountTypeSavings    AccountType = "savings"
	AccountTypeInvestment AccountType = "investment"
)

type Account struct {
	ID          uuid.UUID
	Name        string
	Type        AccountType
	Balance     valueobject.Money
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func NewAccount(name string, accountType AccountType, initialBalance valueobject.Money, description string) *Account {
	now := time.Now()
	return &Account{
		ID:          uuid.New(),
		Name:        name,
		Type:        accountType,
		Balance:     initialBalance,
		Description: description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func (a *Account) Deposit(amount valueobject.Money) error {
	newBalance, err := a.Balance.Add(amount)
	if err != nil {
		return err
	}
	a.Balance = newBalance
	a.UpdatedAt = time.Now()
	return nil
}

func (a *Account) Withdraw(amount valueobject.Money) error {
	newBalance, err := a.Balance.Subtract(amount)
	if err != nil {
		return err
	}
	if newBalance.IsNegative() && a.Type != AccountTypeChecking {
		return fmt.Errorf("insufficient funds: balance would be %s", newBalance.String())
	}
	a.Balance = newBalance
	a.UpdatedAt = time.Now()
	return nil
}

func (a *Account) GetAvailableBalance() valueobject.Money {
	return a.Balance
}

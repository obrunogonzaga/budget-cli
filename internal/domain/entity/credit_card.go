package entity

import (
	"fmt"
	"time"

	"financli/internal/domain/valueobject"
	"github.com/google/uuid"
)

type CreditCard struct {
	ID             uuid.UUID
	AccountID      uuid.UUID
	Name           string
	LastFourDigits string
	CreditLimit    valueobject.Money
	CurrentBalance valueobject.Money
	DueDay         int
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func NewCreditCard(accountID uuid.UUID, name string, lastFourDigits string, creditLimit valueobject.Money, dueDay int) (*CreditCard, error) {
	if dueDay < 1 || dueDay > 31 {
		return nil, fmt.Errorf("due day must be between 1 and 31")
	}

	if len(lastFourDigits) != 4 {
		return nil, fmt.Errorf("last four digits must be exactly 4 characters")
	}

	now := time.Now()
	return &CreditCard{
		ID:             uuid.New(),
		AccountID:      accountID,
		Name:           name,
		LastFourDigits: lastFourDigits,
		CreditLimit:    creditLimit,
		CurrentBalance: valueobject.NewMoney(0, creditLimit.Currency()),
		DueDay:         dueDay,
		CreatedAt:      now,
		UpdatedAt:      now,
	}, nil
}

func (c *CreditCard) Charge(amount valueobject.Money) error {
	newBalance, err := c.CurrentBalance.Add(amount)
	if err != nil {
		return err
	}

	availableCredit, err := c.CreditLimit.Subtract(newBalance)
	if err != nil {
		return err
	}

	if availableCredit.IsNegative() {
		return fmt.Errorf("credit limit exceeded: limit is %s, would be %s", c.CreditLimit.String(), newBalance.String())
	}

	c.CurrentBalance = newBalance
	c.UpdatedAt = time.Now()
	return nil
}

func (c *CreditCard) Payment(amount valueobject.Money) error {
	newBalance, err := c.CurrentBalance.Subtract(amount)
	if err != nil {
		return err
	}

	if newBalance.IsNegative() {
		c.CurrentBalance = valueobject.NewMoney(0, c.CurrentBalance.Currency())
	} else {
		c.CurrentBalance = newBalance
	}

	c.UpdatedAt = time.Now()
	return nil
}

func (c *CreditCard) GetAvailableCredit() (valueobject.Money, error) {
	return c.CreditLimit.Subtract(c.CurrentBalance)
}

func (c *CreditCard) GetUtilizationPercentage() float64 {
	if c.CreditLimit.IsZero() {
		return 0
	}
	return (c.CurrentBalance.Amount() / c.CreditLimit.Amount()) * 100
}

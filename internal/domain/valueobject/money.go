package valueobject

import (
	"fmt"
	"math"
)

type Money struct {
	amount   float64
	currency string
}

func NewMoney(amount float64, currency string) Money {
	return Money{
		amount:   math.Round(amount*100) / 100,
		currency: currency,
	}
}

func (m Money) Amount() float64 {
	return m.amount
}

func (m Money) Currency() string {
	return m.currency
}

func (m Money) Add(other Money) (Money, error) {
	if m.currency != other.currency {
		return Money{}, fmt.Errorf("cannot add different currencies: %s and %s", m.currency, other.currency)
	}
	return NewMoney(m.amount+other.amount, m.currency), nil
}

func (m Money) Subtract(other Money) (Money, error) {
	if m.currency != other.currency {
		return Money{}, fmt.Errorf("cannot subtract different currencies: %s and %s", m.currency, other.currency)
	}
	return NewMoney(m.amount-other.amount, m.currency), nil
}

func (m Money) Multiply(factor float64) Money {
	return NewMoney(m.amount*factor, m.currency)
}

func (m Money) String() string {
	if m.currency == "BRL" {
		return fmt.Sprintf("R$ %.2f", m.amount)
	}
	return fmt.Sprintf("%s %.2f", m.currency, m.amount)
}

func (m Money) IsNegative() bool {
	return m.amount < 0
}

func (m Money) IsZero() bool {
	return m.amount == 0
}

func (m Money) Equals(other Money) bool {
	return m.amount == other.amount && m.currency == other.currency
}

package entity

import (
	"testing"

	"financli/internal/domain/valueobject"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAccount(t *testing.T) {
	money := valueobject.NewMoney(1000.0, "BRL")
	account := NewAccount("Test Account", AccountTypeChecking, money, "Test description")

	assert.Equal(t, "Test Account", account.Name)
	assert.Equal(t, AccountTypeChecking, account.Type)
	assert.Equal(t, money, account.Balance)
	assert.Equal(t, "Test description", account.Description)
	assert.NotEmpty(t, account.ID)
}

func TestAccount_Deposit(t *testing.T) {
	initialBalance := valueobject.NewMoney(1000.0, "BRL")
	account := NewAccount("Test Account", AccountTypeChecking, initialBalance, "Test")

	depositAmount := valueobject.NewMoney(500.0, "BRL")
	err := account.Deposit(depositAmount)

	require.NoError(t, err)
	assert.Equal(t, 1500.0, account.Balance.Amount())
}

func TestAccount_Withdraw(t *testing.T) {
	initialBalance := valueobject.NewMoney(1000.0, "BRL")
	account := NewAccount("Test Account", AccountTypeChecking, initialBalance, "Test")

	withdrawAmount := valueobject.NewMoney(300.0, "BRL")
	err := account.Withdraw(withdrawAmount)

	require.NoError(t, err)
	assert.Equal(t, 700.0, account.Balance.Amount())
}

func TestAccount_WithdrawInsufficientFunds(t *testing.T) {
	initialBalance := valueobject.NewMoney(1000.0, "BRL")
	account := NewAccount("Test Account", AccountTypeSavings, initialBalance, "Test")

	withdrawAmount := valueobject.NewMoney(1500.0, "BRL")
	err := account.Withdraw(withdrawAmount)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "insufficient funds")
	assert.Equal(t, 1000.0, account.Balance.Amount())
}

func TestAccount_CheckingAllowsNegativeBalance(t *testing.T) {
	initialBalance := valueobject.NewMoney(1000.0, "BRL")
	account := NewAccount("Test Account", AccountTypeChecking, initialBalance, "Test")

	withdrawAmount := valueobject.NewMoney(1500.0, "BRL")
	err := account.Withdraw(withdrawAmount)

	require.NoError(t, err)
	assert.Equal(t, -500.0, account.Balance.Amount())
}

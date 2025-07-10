package usecase

import (
	"context"
	"fmt"

	"financli/internal/domain/entity"
	"financli/internal/domain/repository"
	"financli/internal/domain/valueobject"
	"github.com/google/uuid"
)

type AccountUseCase struct {
	accountRepo repository.AccountRepository
}

func NewAccountUseCase(accountRepo repository.AccountRepository) *AccountUseCase {
	return &AccountUseCase{
		accountRepo: accountRepo,
	}
}

func (uc *AccountUseCase) CreateAccount(ctx context.Context, name string, accountType entity.AccountType, initialBalance float64, currency, description string) (*entity.Account, error) {
	money := valueobject.NewMoney(initialBalance, currency)
	account := entity.NewAccount(name, accountType, money, description)

	if err := uc.accountRepo.Create(ctx, account); err != nil {
		return nil, fmt.Errorf("failed to create account: %w", err)
	}

	return account, nil
}

func (uc *AccountUseCase) GetAccount(ctx context.Context, id uuid.UUID) (*entity.Account, error) {
	return uc.accountRepo.FindByID(ctx, id)
}

func (uc *AccountUseCase) ListAccounts(ctx context.Context) ([]*entity.Account, error) {
	return uc.accountRepo.FindAll(ctx)
}

func (uc *AccountUseCase) Deposit(ctx context.Context, accountID uuid.UUID, amount float64, currency string) error {
	account, err := uc.accountRepo.FindByID(ctx, accountID)
	if err != nil {
		return err
	}

	money := valueobject.NewMoney(amount, currency)
	if err := account.Deposit(money); err != nil {
		return err
	}

	return uc.accountRepo.Update(ctx, account)
}

func (uc *AccountUseCase) Withdraw(ctx context.Context, accountID uuid.UUID, amount float64, currency string) error {
	account, err := uc.accountRepo.FindByID(ctx, accountID)
	if err != nil {
		return err
	}

	money := valueobject.NewMoney(amount, currency)
	if err := account.Withdraw(money); err != nil {
		return err
	}

	return uc.accountRepo.Update(ctx, account)
}

func (uc *AccountUseCase) UpdateAccount(ctx context.Context, id uuid.UUID, name string, accountType entity.AccountType, balance float64, currency, description string) (*entity.Account, error) {
	account, err := uc.accountRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("account not found: %w", err)
	}

	account.Name = name
	account.Type = accountType
	account.Description = description
	account.Balance = valueobject.NewMoney(balance, currency)

	if err := uc.accountRepo.Update(ctx, account); err != nil {
		return nil, fmt.Errorf("failed to update account: %w", err)
	}

	return account, nil
}

func (uc *AccountUseCase) DeleteAccount(ctx context.Context, id uuid.UUID) error {
	return uc.accountRepo.Delete(ctx, id)
}

func (uc *AccountUseCase) Transfer(ctx context.Context, fromAccountID, toAccountID uuid.UUID, amount float64, currency string) error {
	fromAccount, err := uc.accountRepo.FindByID(ctx, fromAccountID)
	if err != nil {
		return fmt.Errorf("source account not found: %w", err)
	}

	toAccount, err := uc.accountRepo.FindByID(ctx, toAccountID)
	if err != nil {
		return fmt.Errorf("destination account not found: %w", err)
	}

	money := valueobject.NewMoney(amount, currency)

	if err := fromAccount.Withdraw(money); err != nil {
		return fmt.Errorf("failed to withdraw from source account: %w", err)
	}

	if err := toAccount.Deposit(money); err != nil {
		return fmt.Errorf("failed to deposit to destination account: %w", err)
	}

	if err := uc.accountRepo.Update(ctx, fromAccount); err != nil {
		return fmt.Errorf("failed to update source account: %w", err)
	}

	if err := uc.accountRepo.Update(ctx, toAccount); err != nil {
		return fmt.Errorf("failed to update destination account: %w", err)
	}

	return nil
}

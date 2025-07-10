package usecase

import (
	"context"
	"fmt"

	"financli/internal/domain/entity"
	"financli/internal/domain/repository"
	"financli/internal/domain/valueobject"
	"github.com/google/uuid"
)

type CreditCardUseCase struct {
	creditCardRepo repository.CreditCardRepository
	accountRepo    repository.AccountRepository
}

func NewCreditCardUseCase(creditCardRepo repository.CreditCardRepository, accountRepo repository.AccountRepository) *CreditCardUseCase {
	return &CreditCardUseCase{
		creditCardRepo: creditCardRepo,
		accountRepo:    accountRepo,
	}
}

func (uc *CreditCardUseCase) CreateCreditCard(ctx context.Context, accountID uuid.UUID, name, lastFourDigits string, creditLimit float64, currency string, dueDay int) (*entity.CreditCard, error) {
	// Verify account exists
	_, err := uc.accountRepo.FindByID(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("account not found: %w", err)
	}

	limit := valueobject.NewMoney(creditLimit, currency)
	card, err := entity.NewCreditCard(accountID, name, lastFourDigits, limit, dueDay)
	if err != nil {
		return nil, err
	}

	if err := uc.creditCardRepo.Create(ctx, card); err != nil {
		return nil, fmt.Errorf("failed to create credit card: %w", err)
	}

	return card, nil
}

func (uc *CreditCardUseCase) GetCreditCard(ctx context.Context, id uuid.UUID) (*entity.CreditCard, error) {
	return uc.creditCardRepo.FindByID(ctx, id)
}

func (uc *CreditCardUseCase) ListCreditCards(ctx context.Context) ([]*entity.CreditCard, error) {
	return uc.creditCardRepo.FindAll(ctx)
}

func (uc *CreditCardUseCase) ListCreditCardsByAccount(ctx context.Context, accountID uuid.UUID) ([]*entity.CreditCard, error) {
	return uc.creditCardRepo.FindByAccountID(ctx, accountID)
}

func (uc *CreditCardUseCase) ChargeCard(ctx context.Context, cardID uuid.UUID, amount float64, currency string) error {
	card, err := uc.creditCardRepo.FindByID(ctx, cardID)
	if err != nil {
		return err
	}

	money := valueobject.NewMoney(amount, currency)
	if err := card.Charge(money); err != nil {
		return err
	}

	return uc.creditCardRepo.Update(ctx, card)
}

func (uc *CreditCardUseCase) MakePayment(ctx context.Context, cardID uuid.UUID, amount float64, currency string) error {
	card, err := uc.creditCardRepo.FindByID(ctx, cardID)
	if err != nil {
		return err
	}

	account, err := uc.accountRepo.FindByID(ctx, card.AccountID)
	if err != nil {
		return fmt.Errorf("linked account not found: %w", err)
	}

	money := valueobject.NewMoney(amount, currency)

	// Withdraw from account
	if err := account.Withdraw(money); err != nil {
		return fmt.Errorf("failed to withdraw from account: %w", err)
	}

	// Apply payment to card
	if err := card.Payment(money); err != nil {
		return fmt.Errorf("failed to apply payment to card: %w", err)
	}

	// Update both
	if err := uc.accountRepo.Update(ctx, account); err != nil {
		return fmt.Errorf("failed to update account: %w", err)
	}

	if err := uc.creditCardRepo.Update(ctx, card); err != nil {
		return fmt.Errorf("failed to update credit card: %w", err)
	}

	return nil
}

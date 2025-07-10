package repository

import (
	"context"

	"financli/internal/domain/entity"
	"github.com/google/uuid"
)

type CreditCardRepository interface {
	Create(ctx context.Context, card *entity.CreditCard) error
	Update(ctx context.Context, card *entity.CreditCard) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.CreditCard, error)
	FindAll(ctx context.Context) ([]*entity.CreditCard, error)
	FindByAccountID(ctx context.Context, accountID uuid.UUID) ([]*entity.CreditCard, error)
}

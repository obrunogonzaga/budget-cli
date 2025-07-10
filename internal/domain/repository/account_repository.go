package repository

import (
	"context"

	"financli/internal/domain/entity"
	"github.com/google/uuid"
)

type AccountRepository interface {
	Create(ctx context.Context, account *entity.Account) error
	Update(ctx context.Context, account *entity.Account) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Account, error)
	FindAll(ctx context.Context) ([]*entity.Account, error)
	FindByType(ctx context.Context, accountType entity.AccountType) ([]*entity.Account, error)
}

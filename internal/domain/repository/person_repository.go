package repository

import (
	"context"

	"financli/internal/domain/entity"
	"github.com/google/uuid"
)

type PersonRepository interface {
	Create(ctx context.Context, person *entity.Person) error
	Update(ctx context.Context, person *entity.Person) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Person, error)
	FindAll(ctx context.Context) ([]*entity.Person, error)
	FindByEmail(ctx context.Context, email string) (*entity.Person, error)
}

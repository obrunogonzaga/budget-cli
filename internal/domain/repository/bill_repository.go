package repository

import (
	"context"
	"time"

	"financli/internal/domain/entity"
	"github.com/google/uuid"
)

type BillRepository interface {
	Create(ctx context.Context, bill *entity.Bill) error
	Update(ctx context.Context, bill *entity.Bill) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Bill, error)
	FindAll(ctx context.Context) ([]*entity.Bill, error)
	FindByStatus(ctx context.Context, status entity.BillStatus) ([]*entity.Bill, error)
	FindByDateRange(ctx context.Context, startDate, endDate time.Time) ([]*entity.Bill, error)
	FindOverdue(ctx context.Context) ([]*entity.Bill, error)
}
package repository

import (
	"context"
	"time"

	"financli/internal/domain/entity"
	"github.com/google/uuid"
)

type TransactionRepository interface {
	Create(ctx context.Context, transaction *entity.Transaction) error
	Update(ctx context.Context, transaction *entity.Transaction) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.Transaction, error)
	FindAll(ctx context.Context) ([]*entity.Transaction, error)
	FindByAccountID(ctx context.Context, accountID uuid.UUID) ([]*entity.Transaction, error)
	FindByCreditCardID(ctx context.Context, creditCardID uuid.UUID) ([]*entity.Transaction, error)
	FindByCreditCardInvoiceID(ctx context.Context, invoiceID uuid.UUID) ([]*entity.Transaction, error)
	FindByBillID(ctx context.Context, billID uuid.UUID) ([]*entity.Transaction, error)
	FindByDateRange(ctx context.Context, startDate, endDate time.Time) ([]*entity.Transaction, error)
	FindByCategory(ctx context.Context, category entity.TransactionCategory) ([]*entity.Transaction, error)
	FindSharedWithPerson(ctx context.Context, personID uuid.UUID) ([]*entity.Transaction, error)
	FindUnassignedToBill(ctx context.Context, startDate, endDate time.Time) ([]*entity.Transaction, error)
}

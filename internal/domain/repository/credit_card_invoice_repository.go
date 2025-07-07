package repository

import (
	"context"
	"time"

	"financli/internal/domain/entity"
	"github.com/google/uuid"
)

type CreditCardInvoiceRepository interface {
	Create(ctx context.Context, invoice *entity.CreditCardInvoice) error
	Update(ctx context.Context, invoice *entity.CreditCardInvoice) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.CreditCardInvoice, error)
	FindByCreditCard(ctx context.Context, creditCardID uuid.UUID) ([]*entity.CreditCardInvoice, error)
	FindByMonth(ctx context.Context, creditCardID uuid.UUID, referenceMonth string) (*entity.CreditCardInvoice, error)
	FindOpenInvoice(ctx context.Context, creditCardID uuid.UUID) (*entity.CreditCardInvoice, error)
	FindByDateRange(ctx context.Context, creditCardID uuid.UUID, startDate, endDate time.Time) ([]*entity.CreditCardInvoice, error)
	FindByStatus(ctx context.Context, creditCardID uuid.UUID, status entity.InvoiceStatus) ([]*entity.CreditCardInvoice, error)
}
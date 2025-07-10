package usecase

import (
	"context"
	"fmt"
	"time"

	"financli/internal/domain/entity"
	"financli/internal/domain/repository"
	"financli/internal/domain/valueobject"
	"github.com/google/uuid"
)

type BillUseCase struct {
	billRepo repository.BillRepository
}

func NewBillUseCase(billRepo repository.BillRepository) *BillUseCase {
	return &BillUseCase{
		billRepo: billRepo,
	}
}

func (uc *BillUseCase) CreateBill(ctx context.Context, name, description string, startDate, endDate, dueDate time.Time, totalAmount float64, currency string) (*entity.Bill, error) {
	money := valueobject.NewMoney(totalAmount, currency)
	bill, err := entity.NewBill(name, description, startDate, endDate, dueDate, money)
	if err != nil {
		return nil, err
	}

	if err := uc.billRepo.Create(ctx, bill); err != nil {
		return nil, fmt.Errorf("failed to create bill: %w", err)
	}

	return bill, nil
}

func (uc *BillUseCase) GetBill(ctx context.Context, id uuid.UUID) (*entity.Bill, error) {
	return uc.billRepo.FindByID(ctx, id)
}

func (uc *BillUseCase) ListBills(ctx context.Context) ([]*entity.Bill, error) {
	return uc.billRepo.FindAll(ctx)
}

func (uc *BillUseCase) GetBillsByStatus(ctx context.Context, status entity.BillStatus) ([]*entity.Bill, error) {
	return uc.billRepo.FindByStatus(ctx, status)
}

func (uc *BillUseCase) GetPendingBills(ctx context.Context) ([]*entity.Bill, error) {
	return uc.billRepo.FindByStatus(ctx, entity.BillStatusOpen)
}

func (uc *BillUseCase) GetOverdueBills(ctx context.Context) ([]*entity.Bill, error) {
	return uc.billRepo.FindOverdue(ctx)
}

func (uc *BillUseCase) AddPayment(ctx context.Context, billID uuid.UUID, amount float64, currency string) error {
	bill, err := uc.billRepo.FindByID(ctx, billID)
	if err != nil {
		return err
	}

	money := valueobject.NewMoney(amount, currency)
	if err := bill.AddPayment(money); err != nil {
		return err
	}

	return uc.billRepo.Update(ctx, bill)
}

func (uc *BillUseCase) CloseBill(ctx context.Context, billID uuid.UUID) error {
	bill, err := uc.billRepo.FindByID(ctx, billID)
	if err != nil {
		return err
	}

	if err := bill.Close(); err != nil {
		return err
	}

	return uc.billRepo.Update(ctx, bill)
}

func (uc *BillUseCase) GetBillsByDateRange(ctx context.Context, startDate, endDate time.Time) ([]*entity.Bill, error) {
	return uc.billRepo.FindByDateRange(ctx, startDate, endDate)
}

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

type ReportUseCase struct {
	transactionRepo repository.TransactionRepository
	personRepo      repository.PersonRepository
	billRepo        repository.BillRepository
}

type SharedExpenseReport struct {
	Person      *entity.Person
	TotalOwed   valueobject.Money
	TotalPaid   valueobject.Money
	Balance     valueobject.Money
	Expenses    []*entity.Transaction
}

type BillReport struct {
	Bill            *entity.Bill
	TotalExpenses   valueobject.Money
	SharedExpenses  valueobject.Money
	PersonalExpenses valueobject.Money
	Participants    []string
}

func NewReportUseCase(
	transactionRepo repository.TransactionRepository,
	personRepo repository.PersonRepository,
	billRepo repository.BillRepository,
) *ReportUseCase {
	return &ReportUseCase{
		transactionRepo: transactionRepo,
		personRepo:      personRepo,
		billRepo:        billRepo,
	}
}

func (uc *ReportUseCase) GetSharedExpenseReport(ctx context.Context, personID uuid.UUID, startDate, endDate time.Time) (*SharedExpenseReport, error) {
	person, err := uc.personRepo.FindByID(ctx, personID)
	if err != nil {
		return nil, fmt.Errorf("person not found: %w", err)
	}
	
	transactions, err := uc.transactionRepo.FindSharedWithPerson(ctx, personID)
	if err != nil {
		return nil, fmt.Errorf("failed to get shared transactions: %w", err)
	}
	
	var filteredTransactions []*entity.Transaction
	totalOwed := valueobject.NewMoney(0, "BRL")
	
	for _, txn := range transactions {
		if txn.Date.After(startDate) && txn.Date.Before(endDate) {
			filteredTransactions = append(filteredTransactions, txn)
			
			for _, shared := range txn.SharedWith {
				if shared.PersonID == personID {
					owed, err := totalOwed.Add(shared.Amount)
					if err == nil {
						totalOwed = owed
					}
				}
			}
		}
	}
	
	// For simplicity, assuming no payments made yet
	totalPaid := valueobject.NewMoney(0, totalOwed.Currency())
	balance := totalOwed
	
	return &SharedExpenseReport{
		Person:    person,
		TotalOwed: totalOwed,
		TotalPaid: totalPaid,
		Balance:   balance,
		Expenses:  filteredTransactions,
	}, nil
}

func (uc *ReportUseCase) GetBillReport(ctx context.Context, billID uuid.UUID) (*BillReport, error) {
	bill, err := uc.billRepo.FindByID(ctx, billID)
	if err != nil {
		return nil, fmt.Errorf("bill not found: %w", err)
	}
	
	transactions, err := uc.transactionRepo.FindByBillID(ctx, billID)
	if err != nil {
		return nil, fmt.Errorf("failed to get bill transactions: %w", err)
	}
	
	totalExpenses := valueobject.NewMoney(0, "BRL")
	sharedExpenses := valueobject.NewMoney(0, "BRL")
	personalExpenses := valueobject.NewMoney(0, "BRL")
	participantMap := make(map[string]bool)
	
	for _, txn := range transactions {
		total, err := totalExpenses.Add(txn.Amount)
		if err == nil {
			totalExpenses = total
		}
		
		personalAmount := txn.GetPersonalAmount()
		personal, err := personalExpenses.Add(personalAmount)
		if err == nil {
			personalExpenses = personal
		}
		
		for _, shared := range txn.SharedWith {
			sharedAmount, err := sharedExpenses.Add(shared.Amount)
			if err == nil {
				sharedExpenses = sharedAmount
			}
			
			// Get participant name
			person, err := uc.personRepo.FindByID(ctx, shared.PersonID)
			if err == nil {
				participantMap[person.Name] = true
			}
		}
	}
	
	var participants []string
	for name := range participantMap {
		participants = append(participants, name)
	}
	
	return &BillReport{
		Bill:             bill,
		TotalExpenses:    totalExpenses,
		SharedExpenses:   sharedExpenses,
		PersonalExpenses: personalExpenses,
		Participants:     participants,
	}, nil
}

func (uc *ReportUseCase) GetMonthlyReport(ctx context.Context, year int, month time.Month) (map[string]interface{}, error) {
	startDate := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, 0)
	
	transactions, err := uc.transactionRepo.FindByDateRange(ctx, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get transactions: %w", err)
	}
	
	totalIncome := valueobject.NewMoney(0, "BRL")
	totalExpenses := valueobject.NewMoney(0, "BRL")
	categoryBreakdown := make(map[entity.TransactionCategory]valueobject.Money)
	
	for _, txn := range transactions {
		if txn.Type == entity.TransactionTypeCredit {
			income, err := totalIncome.Add(txn.Amount)
			if err == nil {
				totalIncome = income
			}
		} else {
			expenses, err := totalExpenses.Add(txn.Amount)
			if err == nil {
				totalExpenses = expenses
			}
		}
		
		if existing, exists := categoryBreakdown[txn.Category]; exists {
			updated, err := existing.Add(txn.Amount)
			if err == nil {
				categoryBreakdown[txn.Category] = updated
			}
		} else {
			categoryBreakdown[txn.Category] = txn.Amount
		}
	}
	
	return map[string]interface{}{
		"period":             fmt.Sprintf("%s %d", month.String(), year),
		"totalIncome":        totalIncome,
		"totalExpenses":      totalExpenses,
		"netSavings":         func() valueobject.Money { net, _ := totalIncome.Subtract(totalExpenses); return net }(),
		"categoryBreakdown":  categoryBreakdown,
		"transactionCount":   len(transactions),
	}, nil
}
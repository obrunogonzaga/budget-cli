package mongodb

import (
	"financli/internal/domain/entity"
	"financli/internal/domain/valueobject"
	"github.com/google/uuid"
)

func MoneyToModel(money valueobject.Money) MoneyModel {
	return MoneyModel{
		Amount:   money.Amount(),
		Currency: money.Currency(),
	}
}

func MoneyFromModel(model MoneyModel) valueobject.Money {
	return valueobject.NewMoney(model.Amount, model.Currency)
}

func AccountToModel(account *entity.Account) AccountModel {
	return AccountModel{
		UUID:        account.ID.String(),
		Name:        account.Name,
		Type:        string(account.Type),
		Balance:     MoneyToModel(account.Balance),
		Description: account.Description,
		CreatedAt:   account.CreatedAt,
		UpdatedAt:   account.UpdatedAt,
	}
}

func AccountFromModel(model AccountModel) (*entity.Account, error) {
	id, err := uuid.Parse(model.UUID)
	if err != nil {
		return nil, err
	}

	return &entity.Account{
		ID:          id,
		Name:        model.Name,
		Type:        entity.AccountType(model.Type),
		Balance:     MoneyFromModel(model.Balance),
		Description: model.Description,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
	}, nil
}

func CreditCardToModel(card *entity.CreditCard) CreditCardModel {
	return CreditCardModel{
		UUID:           card.ID.String(),
		AccountUUID:    card.AccountID.String(),
		Name:           card.Name,
		LastFourDigits: card.LastFourDigits,
		CreditLimit:    MoneyToModel(card.CreditLimit),
		CurrentBalance: MoneyToModel(card.CurrentBalance),
		DueDay:         card.DueDay,
		CreatedAt:      card.CreatedAt,
		UpdatedAt:      card.UpdatedAt,
	}
}

func CreditCardFromModel(model CreditCardModel) (*entity.CreditCard, error) {
	id, err := uuid.Parse(model.UUID)
	if err != nil {
		return nil, err
	}

	accountID, err := uuid.Parse(model.AccountUUID)
	if err != nil {
		return nil, err
	}

	return &entity.CreditCard{
		ID:             id,
		AccountID:      accountID,
		Name:           model.Name,
		LastFourDigits: model.LastFourDigits,
		CreditLimit:    MoneyFromModel(model.CreditLimit),
		CurrentBalance: MoneyFromModel(model.CurrentBalance),
		DueDay:         model.DueDay,
		CreatedAt:      model.CreatedAt,
		UpdatedAt:      model.UpdatedAt,
	}, nil
}

func PersonToModel(person *entity.Person) PersonModel {
	return PersonModel{
		UUID:      person.ID.String(),
		Name:      person.Name,
		Email:     person.Email,
		Phone:     person.Phone,
		CreatedAt: person.CreatedAt,
		UpdatedAt: person.UpdatedAt,
	}
}

func PersonFromModel(model PersonModel) (*entity.Person, error) {
	id, err := uuid.Parse(model.UUID)
	if err != nil {
		return nil, err
	}

	return &entity.Person{
		ID:        id,
		Name:      model.Name,
		Email:     model.Email,
		Phone:     model.Phone,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
	}, nil
}

func BillToModel(bill *entity.Bill) BillModel {
	return BillModel{
		UUID:        bill.ID.String(),
		Name:        bill.Name,
		Description: bill.Description,
		StartDate:   bill.StartDate,
		EndDate:     bill.EndDate,
		DueDate:     bill.DueDate,
		TotalAmount: MoneyToModel(bill.TotalAmount),
		PaidAmount:  MoneyToModel(bill.PaidAmount),
		Status:      string(bill.Status),
		CreatedAt:   bill.CreatedAt,
		UpdatedAt:   bill.UpdatedAt,
	}
}

func BillFromModel(model BillModel) (*entity.Bill, error) {
	id, err := uuid.Parse(model.UUID)
	if err != nil {
		return nil, err
	}

	return &entity.Bill{
		ID:          id,
		Name:        model.Name,
		Description: model.Description,
		StartDate:   model.StartDate,
		EndDate:     model.EndDate,
		DueDate:     model.DueDate,
		TotalAmount: MoneyFromModel(model.TotalAmount),
		PaidAmount:  MoneyFromModel(model.PaidAmount),
		Status:      entity.BillStatus(model.Status),
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
	}, nil
}

func TransactionToModel(transaction *entity.Transaction) TransactionModel {
	model := TransactionModel{
		UUID:        transaction.ID.String(),
		Type:        string(transaction.Type),
		Category:    string(transaction.Category),
		Amount:      MoneyToModel(transaction.Amount),
		Description: transaction.Description,
		Date:        transaction.Date,
		SharedWith:  make([]SharedExpenseModel, len(transaction.SharedWith)),
		CreatedAt:   transaction.CreatedAt,
		UpdatedAt:   transaction.UpdatedAt,
	}

	if transaction.AccountID != nil {
		accountUUID := transaction.AccountID.String()
		model.AccountUUID = &accountUUID
	}

	if transaction.CreditCardID != nil {
		creditCardUUID := transaction.CreditCardID.String()
		model.CreditCardUUID = &creditCardUUID
	}

	if transaction.BillID != nil {
		billUUID := transaction.BillID.String()
		model.BillUUID = &billUUID
	}

	for i, shared := range transaction.SharedWith {
		model.SharedWith[i] = SharedExpenseModel{
			PersonUUID: shared.PersonID.String(),
			Amount:     MoneyToModel(shared.Amount),
			Percentage: shared.Percentage,
		}
	}

	return model
}

func TransactionFromModel(model TransactionModel) (*entity.Transaction, error) {
	id, err := uuid.Parse(model.UUID)
	if err != nil {
		return nil, err
	}

	transaction := &entity.Transaction{
		ID:          id,
		Type:        entity.TransactionType(model.Type),
		Category:    entity.TransactionCategory(model.Category),
		Amount:      MoneyFromModel(model.Amount),
		Description: model.Description,
		Date:        model.Date,
		SharedWith:  make([]entity.SharedExpense, len(model.SharedWith)),
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
	}

	if model.AccountUUID != nil {
		accountID, err := uuid.Parse(*model.AccountUUID)
		if err != nil {
			return nil, err
		}
		transaction.AccountID = &accountID
	}

	if model.CreditCardUUID != nil {
		creditCardID, err := uuid.Parse(*model.CreditCardUUID)
		if err != nil {
			return nil, err
		}
		transaction.CreditCardID = &creditCardID
	}

	if model.BillUUID != nil {
		billID, err := uuid.Parse(*model.BillUUID)
		if err != nil {
			return nil, err
		}
		transaction.BillID = &billID
	}

	for i, shared := range model.SharedWith {
		personID, err := uuid.Parse(shared.PersonUUID)
		if err != nil {
			return nil, err
		}
		transaction.SharedWith[i] = entity.SharedExpense{
			PersonID:   personID,
			Amount:     MoneyFromModel(shared.Amount),
			Percentage: shared.Percentage,
		}
	}

	return transaction, nil
}
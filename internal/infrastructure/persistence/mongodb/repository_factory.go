package mongodb

import (
	"context"
	"time"

	"financli/internal/domain/entity"
	"financli/internal/domain/repository"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// Simple implementations for remaining repositories

type creditCardRepository struct {
	collection *mongo.Collection
}

func NewCreditCardRepository(db *mongo.Database) repository.CreditCardRepository {
	return &creditCardRepository{
		collection: db.Collection("credit_cards"),
	}
}

func (r *creditCardRepository) Create(ctx context.Context, card *entity.CreditCard) error {
	model := CreditCardToModel(card)
	_, err := r.collection.InsertOne(ctx, model)
	return err
}

func (r *creditCardRepository) Update(ctx context.Context, card *entity.CreditCard) error {
	model := CreditCardToModel(card)
	filter := bson.M{"uuid": card.ID.String()}
	update := bson.M{"$set": model}
	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

func (r *creditCardRepository) Delete(ctx context.Context, id uuid.UUID) error {
	filter := bson.M{"uuid": id.String()}
	_, err := r.collection.DeleteOne(ctx, filter)
	return err
}

func (r *creditCardRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.CreditCard, error) {
	var model CreditCardModel
	filter := bson.M{"uuid": id.String()}
	err := r.collection.FindOne(ctx, filter).Decode(&model)
	if err != nil {
		return nil, err
	}
	return CreditCardFromModel(model)
}

func (r *creditCardRepository) FindAll(ctx context.Context) ([]*entity.CreditCard, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var cards []*entity.CreditCard
	for cursor.Next(ctx) {
		var model CreditCardModel
		if err := cursor.Decode(&model); err != nil {
			return nil, err
		}
		card, err := CreditCardFromModel(model)
		if err != nil {
			return nil, err
		}
		cards = append(cards, card)
	}
	return cards, nil
}

func (r *creditCardRepository) FindByAccountID(ctx context.Context, accountID uuid.UUID) ([]*entity.CreditCard, error) {
	filter := bson.M{"account_uuid": accountID.String()}
	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var cards []*entity.CreditCard
	for cursor.Next(ctx) {
		var model CreditCardModel
		if err := cursor.Decode(&model); err != nil {
			return nil, err
		}
		card, err := CreditCardFromModel(model)
		if err != nil {
			return nil, err
		}
		cards = append(cards, card)
	}
	return cards, nil
}

// PersonRepository implementation
type personRepository struct {
	collection *mongo.Collection
}

func NewPersonRepository(db *mongo.Database) repository.PersonRepository {
	return &personRepository{
		collection: db.Collection("people"),
	}
}

func (r *personRepository) Create(ctx context.Context, person *entity.Person) error {
	model := PersonToModel(person)
	_, err := r.collection.InsertOne(ctx, model)
	return err
}

func (r *personRepository) Update(ctx context.Context, person *entity.Person) error {
	model := PersonToModel(person)
	filter := bson.M{"uuid": person.ID.String()}
	update := bson.M{"$set": model}
	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

func (r *personRepository) Delete(ctx context.Context, id uuid.UUID) error {
	filter := bson.M{"uuid": id.String()}
	_, err := r.collection.DeleteOne(ctx, filter)
	return err
}

func (r *personRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.Person, error) {
	var model PersonModel
	filter := bson.M{"uuid": id.String()}
	err := r.collection.FindOne(ctx, filter).Decode(&model)
	if err != nil {
		return nil, err
	}
	return PersonFromModel(model)
}

func (r *personRepository) FindAll(ctx context.Context) ([]*entity.Person, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var people []*entity.Person
	for cursor.Next(ctx) {
		var model PersonModel
		if err := cursor.Decode(&model); err != nil {
			return nil, err
		}
		person, err := PersonFromModel(model)
		if err != nil {
			return nil, err
		}
		people = append(people, person)
	}
	return people, nil
}

func (r *personRepository) FindByEmail(ctx context.Context, email string) (*entity.Person, error) {
	var model PersonModel
	filter := bson.M{"email": email}
	err := r.collection.FindOne(ctx, filter).Decode(&model)
	if err != nil {
		return nil, err
	}
	return PersonFromModel(model)
}

// BillRepository implementation
type billRepository struct {
	collection *mongo.Collection
}

func NewBillRepository(db *mongo.Database) repository.BillRepository {
	return &billRepository{
		collection: db.Collection("bills"),
	}
}

func (r *billRepository) Create(ctx context.Context, bill *entity.Bill) error {
	model := BillToModel(bill)
	_, err := r.collection.InsertOne(ctx, model)
	return err
}

func (r *billRepository) Update(ctx context.Context, bill *entity.Bill) error {
	model := BillToModel(bill)
	filter := bson.M{"uuid": bill.ID.String()}
	update := bson.M{"$set": model}
	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

func (r *billRepository) Delete(ctx context.Context, id uuid.UUID) error {
	filter := bson.M{"uuid": id.String()}
	_, err := r.collection.DeleteOne(ctx, filter)
	return err
}

func (r *billRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.Bill, error) {
	var model BillModel
	filter := bson.M{"uuid": id.String()}
	err := r.collection.FindOne(ctx, filter).Decode(&model)
	if err != nil {
		return nil, err
	}
	return BillFromModel(model)
}

func (r *billRepository) FindAll(ctx context.Context) ([]*entity.Bill, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var bills []*entity.Bill
	for cursor.Next(ctx) {
		var model BillModel
		if err := cursor.Decode(&model); err != nil {
			return nil, err
		}
		bill, err := BillFromModel(model)
		if err != nil {
			return nil, err
		}
		bills = append(bills, bill)
	}
	return bills, nil
}

func (r *billRepository) FindByStatus(ctx context.Context, status entity.BillStatus) ([]*entity.Bill, error) {
	filter := bson.M{"status": string(status)}
	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var bills []*entity.Bill
	for cursor.Next(ctx) {
		var model BillModel
		if err := cursor.Decode(&model); err != nil {
			return nil, err
		}
		bill, err := BillFromModel(model)
		if err != nil {
			return nil, err
		}
		bills = append(bills, bill)
	}
	return bills, nil
}

func (r *billRepository) FindByDateRange(ctx context.Context, startDate, endDate time.Time) ([]*entity.Bill, error) {
	filter := bson.M{
		"start_date": bson.M{"$lte": endDate},
		"end_date":   bson.M{"$gte": startDate},
	}
	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var bills []*entity.Bill
	for cursor.Next(ctx) {
		var model BillModel
		if err := cursor.Decode(&model); err != nil {
			return nil, err
		}
		bill, err := BillFromModel(model)
		if err != nil {
			return nil, err
		}
		bills = append(bills, bill)
	}
	return bills, nil
}

func (r *billRepository) FindOverdue(ctx context.Context) ([]*entity.Bill, error) {
	filter := bson.M{
		"status":   bson.M{"$ne": string(entity.BillStatusPaid)},
		"due_date": bson.M{"$lt": time.Now()},
	}
	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var bills []*entity.Bill
	for cursor.Next(ctx) {
		var model BillModel
		if err := cursor.Decode(&model); err != nil {
			return nil, err
		}
		bill, err := BillFromModel(model)
		if err != nil {
			return nil, err
		}
		bills = append(bills, bill)
	}
	return bills, nil
}

// TransactionRepository implementation
type transactionRepository struct {
	collection *mongo.Collection
}

func NewTransactionRepository(db *mongo.Database) repository.TransactionRepository {
	return &transactionRepository{
		collection: db.Collection("transactions"),
	}
}

func (r *transactionRepository) Create(ctx context.Context, transaction *entity.Transaction) error {
	model := TransactionToModel(transaction)
	_, err := r.collection.InsertOne(ctx, model)
	return err
}

func (r *transactionRepository) Update(ctx context.Context, transaction *entity.Transaction) error {
	model := TransactionToModel(transaction)
	filter := bson.M{"uuid": transaction.ID.String()}
	update := bson.M{"$set": model}
	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

func (r *transactionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	filter := bson.M{"uuid": id.String()}
	_, err := r.collection.DeleteOne(ctx, filter)
	return err
}

func (r *transactionRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.Transaction, error) {
	var model TransactionModel
	filter := bson.M{"uuid": id.String()}
	err := r.collection.FindOne(ctx, filter).Decode(&model)
	if err != nil {
		return nil, err
	}
	return TransactionFromModel(model)
}

func (r *transactionRepository) FindAll(ctx context.Context) ([]*entity.Transaction, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var transactions []*entity.Transaction
	for cursor.Next(ctx) {
		var model TransactionModel
		if err := cursor.Decode(&model); err != nil {
			return nil, err
		}
		transaction, err := TransactionFromModel(model)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, transaction)
	}
	return transactions, nil
}

func (r *transactionRepository) FindByAccountID(ctx context.Context, accountID uuid.UUID) ([]*entity.Transaction, error) {
	filter := bson.M{"account_uuid": accountID.String()}
	return r.findByFilter(ctx, filter)
}

func (r *transactionRepository) FindByCreditCardID(ctx context.Context, creditCardID uuid.UUID) ([]*entity.Transaction, error) {
	filter := bson.M{"credit_card_uuid": creditCardID.String()}
	return r.findByFilter(ctx, filter)
}

func (r *transactionRepository) FindByBillID(ctx context.Context, billID uuid.UUID) ([]*entity.Transaction, error) {
	filter := bson.M{"bill_uuid": billID.String()}
	return r.findByFilter(ctx, filter)
}

func (r *transactionRepository) FindByCreditCardInvoiceID(ctx context.Context, invoiceID uuid.UUID) ([]*entity.Transaction, error) {
	filter := bson.M{"credit_card_invoice_uuid": invoiceID.String()}
	return r.findByFilter(ctx, filter)
}

func (r *transactionRepository) FindByDateRange(ctx context.Context, startDate, endDate time.Time) ([]*entity.Transaction, error) {
	filter := bson.M{
		"date": bson.M{
			"$gte": startDate,
			"$lte": endDate,
		},
	}
	return r.findByFilter(ctx, filter)
}

func (r *transactionRepository) FindByCategory(ctx context.Context, category entity.TransactionCategory) ([]*entity.Transaction, error) {
	filter := bson.M{"category": string(category)}
	return r.findByFilter(ctx, filter)
}

func (r *transactionRepository) FindSharedWithPerson(ctx context.Context, personID uuid.UUID) ([]*entity.Transaction, error) {
	filter := bson.M{"shared_with.person_uuid": personID.String()}
	return r.findByFilter(ctx, filter)
}

func (r *transactionRepository) FindUnassignedToBill(ctx context.Context, startDate, endDate time.Time) ([]*entity.Transaction, error) {
	filter := bson.M{
		"bill_uuid": bson.M{"$exists": false},
		"date": bson.M{
			"$gte": startDate,
			"$lte": endDate,
		},
	}
	return r.findByFilter(ctx, filter)
}

func (r *transactionRepository) findByFilter(ctx context.Context, filter bson.M) ([]*entity.Transaction, error) {
	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var transactions []*entity.Transaction
	for cursor.Next(ctx) {
		var model TransactionModel
		if err := cursor.Decode(&model); err != nil {
			return nil, err
		}
		transaction, err := TransactionFromModel(model)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, transaction)
	}
	return transactions, nil
}
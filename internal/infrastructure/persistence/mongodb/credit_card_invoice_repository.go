package mongodb

import (
	"context"
	"fmt"
	"time"

	"financli/internal/domain/entity"
	"financli/internal/domain/repository"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type creditCardInvoiceRepository struct {
	collection *mongo.Collection
}

func NewCreditCardInvoiceRepository(db *mongo.Database) repository.CreditCardInvoiceRepository {
	return &creditCardInvoiceRepository{
		collection: db.Collection("credit_card_invoices"),
	}
}

func (r *creditCardInvoiceRepository) Create(ctx context.Context, invoice *entity.CreditCardInvoice) error {
	model := CreditCardInvoiceToModel(invoice)
	_, err := r.collection.InsertOne(ctx, model)
	if err != nil {
		return fmt.Errorf("failed to create credit card invoice: %w", err)
	}
	return nil
}

func (r *creditCardInvoiceRepository) Update(ctx context.Context, invoice *entity.CreditCardInvoice) error {
	model := CreditCardInvoiceToModel(invoice)
	filter := bson.M{"uuid": invoice.ID.String()}
	update := bson.M{"$set": model}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update credit card invoice: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("credit card invoice not found")
	}

	return nil
}

func (r *creditCardInvoiceRepository) Delete(ctx context.Context, id uuid.UUID) error {
	filter := bson.M{"uuid": id.String()}
	result, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to delete credit card invoice: %w", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("credit card invoice not found")
	}

	return nil
}

func (r *creditCardInvoiceRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.CreditCardInvoice, error) {
	var model CreditCardInvoiceModel
	filter := bson.M{"uuid": id.String()}

	err := r.collection.FindOne(ctx, filter).Decode(&model)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("credit card invoice not found")
		}
		return nil, fmt.Errorf("failed to find credit card invoice: %w", err)
	}

	return CreditCardInvoiceFromModel(model)
}

func (r *creditCardInvoiceRepository) FindByCreditCard(ctx context.Context, creditCardID uuid.UUID) ([]*entity.CreditCardInvoice, error) {
	filter := bson.M{"credit_card_uuid": creditCardID.String()}
	opts := options.Find().SetSort(bson.D{{Key: "reference_month", Value: -1}}) // Sort by reference month descending

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find credit card invoices: %w", err)
	}
	defer cursor.Close(ctx)

	var invoices []*entity.CreditCardInvoice
	for cursor.Next(ctx) {
		var model CreditCardInvoiceModel
		if err := cursor.Decode(&model); err != nil {
			return nil, fmt.Errorf("failed to decode credit card invoice: %w", err)
		}

		invoice, err := CreditCardInvoiceFromModel(model)
		if err != nil {
			return nil, err
		}
		invoices = append(invoices, invoice)
	}

	return invoices, nil
}

func (r *creditCardInvoiceRepository) FindByMonth(ctx context.Context, creditCardID uuid.UUID, referenceMonth string) (*entity.CreditCardInvoice, error) {
	var model CreditCardInvoiceModel
	filter := bson.M{
		"credit_card_uuid": creditCardID.String(),
		"reference_month":  referenceMonth,
	}

	err := r.collection.FindOne(ctx, filter).Decode(&model)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("credit card invoice not found for month %s", referenceMonth)
		}
		return nil, fmt.Errorf("failed to find credit card invoice: %w", err)
	}

	return CreditCardInvoiceFromModel(model)
}

func (r *creditCardInvoiceRepository) FindOpenInvoice(ctx context.Context, creditCardID uuid.UUID) (*entity.CreditCardInvoice, error) {
	var model CreditCardInvoiceModel
	filter := bson.M{
		"credit_card_uuid": creditCardID.String(),
		"status":           string(entity.InvoiceStatusOpen),
	}

	err := r.collection.FindOne(ctx, filter).Decode(&model)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("no open credit card invoice found")
		}
		return nil, fmt.Errorf("failed to find open credit card invoice: %w", err)
	}

	return CreditCardInvoiceFromModel(model)
}

func (r *creditCardInvoiceRepository) FindByDateRange(ctx context.Context, creditCardID uuid.UUID, startDate, endDate time.Time) ([]*entity.CreditCardInvoice, error) {
	filter := bson.M{
		"credit_card_uuid": creditCardID.String(),
		"opening_date": bson.M{
			"$gte": startDate,
			"$lte": endDate,
		},
	}
	opts := options.Find().SetSort(bson.D{{Key: "opening_date", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find credit card invoices by date range: %w", err)
	}
	defer cursor.Close(ctx)

	var invoices []*entity.CreditCardInvoice
	for cursor.Next(ctx) {
		var model CreditCardInvoiceModel
		if err := cursor.Decode(&model); err != nil {
			return nil, fmt.Errorf("failed to decode credit card invoice: %w", err)
		}

		invoice, err := CreditCardInvoiceFromModel(model)
		if err != nil {
			return nil, err
		}
		invoices = append(invoices, invoice)
	}

	return invoices, nil
}

func (r *creditCardInvoiceRepository) FindByStatus(ctx context.Context, creditCardID uuid.UUID, status entity.InvoiceStatus) ([]*entity.CreditCardInvoice, error) {
	filter := bson.M{
		"credit_card_uuid": creditCardID.String(),
		"status":           string(status),
	}
	opts := options.Find().SetSort(bson.D{{Key: "due_date", Value: 1}}) // Sort by due date ascending

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find credit card invoices by status: %w", err)
	}
	defer cursor.Close(ctx)

	var invoices []*entity.CreditCardInvoice
	for cursor.Next(ctx) {
		var model CreditCardInvoiceModel
		if err := cursor.Decode(&model); err != nil {
			return nil, fmt.Errorf("failed to decode credit card invoice: %w", err)
		}

		invoice, err := CreditCardInvoiceFromModel(model)
		if err != nil {
			return nil, err
		}
		invoices = append(invoices, invoice)
	}

	return invoices, nil
}

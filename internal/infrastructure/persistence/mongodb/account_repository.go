package mongodb

import (
	"context"
	"fmt"

	"financli/internal/domain/entity"
	"financli/internal/domain/repository"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type accountRepository struct {
	collection *mongo.Collection
}

func NewAccountRepository(db *mongo.Database) repository.AccountRepository {
	return &accountRepository{
		collection: db.Collection("accounts"),
	}
}

func (r *accountRepository) Create(ctx context.Context, account *entity.Account) error {
	model := AccountToModel(account)
	_, err := r.collection.InsertOne(ctx, model)
	if err != nil {
		return fmt.Errorf("failed to create account: %w", err)
	}
	return nil
}

func (r *accountRepository) Update(ctx context.Context, account *entity.Account) error {
	model := AccountToModel(account)
	filter := bson.M{"uuid": account.ID.String()}
	update := bson.M{"$set": model}
	
	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update account: %w", err)
	}
	
	if result.MatchedCount == 0 {
		return fmt.Errorf("account not found")
	}
	
	return nil
}

func (r *accountRepository) Delete(ctx context.Context, id uuid.UUID) error {
	filter := bson.M{"uuid": id.String()}
	result, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to delete account: %w", err)
	}
	
	if result.DeletedCount == 0 {
		return fmt.Errorf("account not found")
	}
	
	return nil
}

func (r *accountRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.Account, error) {
	var model AccountModel
	filter := bson.M{"uuid": id.String()}
	
	err := r.collection.FindOne(ctx, filter).Decode(&model)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("account not found")
		}
		return nil, fmt.Errorf("failed to find account: %w", err)
	}
	
	return AccountFromModel(model)
}

func (r *accountRepository) FindAll(ctx context.Context) ([]*entity.Account, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("failed to find accounts: %w", err)
	}
	defer cursor.Close(ctx)
	
	var accounts []*entity.Account
	for cursor.Next(ctx) {
		var model AccountModel
		if err := cursor.Decode(&model); err != nil {
			return nil, fmt.Errorf("failed to decode account: %w", err)
		}
		
		account, err := AccountFromModel(model)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, account)
	}
	
	return accounts, nil
}

func (r *accountRepository) FindByType(ctx context.Context, accountType entity.AccountType) ([]*entity.Account, error) {
	filter := bson.M{"type": string(accountType)}
	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to find accounts by type: %w", err)
	}
	defer cursor.Close(ctx)
	
	var accounts []*entity.Account
	for cursor.Next(ctx) {
		var model AccountModel
		if err := cursor.Decode(&model); err != nil {
			return nil, fmt.Errorf("failed to decode account: %w", err)
		}
		
		account, err := AccountFromModel(model)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, account)
	}
	
	return accounts, nil
}
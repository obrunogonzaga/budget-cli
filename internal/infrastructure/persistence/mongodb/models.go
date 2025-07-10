package mongodb

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AccountModel struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	UUID        string             `bson:"uuid"`
	Name        string             `bson:"name"`
	Type        string             `bson:"type"`
	Balance     MoneyModel         `bson:"balance"`
	Description string             `bson:"description"`
	CreatedAt   time.Time          `bson:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at"`
}

type CreditCardModel struct {
	ID             primitive.ObjectID `bson:"_id,omitempty"`
	UUID           string             `bson:"uuid"`
	AccountUUID    string             `bson:"account_uuid"`
	Name           string             `bson:"name"`
	LastFourDigits string             `bson:"last_four_digits"`
	CreditLimit    MoneyModel         `bson:"credit_limit"`
	CurrentBalance MoneyModel         `bson:"current_balance"`
	DueDay         int                `bson:"due_day"`
	CreatedAt      time.Time          `bson:"created_at"`
	UpdatedAt      time.Time          `bson:"updated_at"`
}

type PersonModel struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	UUID      string             `bson:"uuid"`
	Name      string             `bson:"name"`
	Email     string             `bson:"email"`
	Phone     string             `bson:"phone"`
	CreatedAt time.Time          `bson:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at"`
}

type BillModel struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	UUID        string             `bson:"uuid"`
	Name        string             `bson:"name"`
	Description string             `bson:"description"`
	StartDate   time.Time          `bson:"start_date"`
	EndDate     time.Time          `bson:"end_date"`
	DueDate     time.Time          `bson:"due_date"`
	TotalAmount MoneyModel         `bson:"total_amount"`
	PaidAmount  MoneyModel         `bson:"paid_amount"`
	Status      string             `bson:"status"`
	CreatedAt   time.Time          `bson:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at"`
}

type TransactionModel struct {
	ID                    primitive.ObjectID   `bson:"_id,omitempty"`
	UUID                  string               `bson:"uuid"`
	AccountUUID           *string              `bson:"account_uuid,omitempty"`
	CreditCardUUID        *string              `bson:"credit_card_uuid,omitempty"`
	CreditCardInvoiceUUID *string              `bson:"credit_card_invoice_uuid,omitempty"`
	BillUUID              *string              `bson:"bill_uuid,omitempty"`
	Type                  string               `bson:"type"`
	Category              string               `bson:"category"`
	Amount                MoneyModel           `bson:"amount"`
	Description           string               `bson:"description"`
	Date                  time.Time            `bson:"date"`
	SharedWith            []SharedExpenseModel `bson:"shared_with"`
	CreatedAt             time.Time            `bson:"created_at"`
	UpdatedAt             time.Time            `bson:"updated_at"`
}

type SharedExpenseModel struct {
	PersonUUID string     `bson:"person_uuid"`
	Amount     MoneyModel `bson:"amount"`
	Percentage float64    `bson:"percentage"`
}

type CreditCardInvoiceModel struct {
	ID               primitive.ObjectID `bson:"_id,omitempty"`
	UUID             string             `bson:"uuid"`
	CreditCardUUID   string             `bson:"credit_card_uuid"`
	ReferenceMonth   string             `bson:"reference_month"`
	OpeningDate      time.Time          `bson:"opening_date"`
	ClosingDate      time.Time          `bson:"closing_date"`
	DueDate          time.Time          `bson:"due_date"`
	PreviousBalance  MoneyModel         `bson:"previous_balance"`
	TotalCharges     MoneyModel         `bson:"total_charges"`
	TotalPayments    MoneyModel         `bson:"total_payments"`
	ClosingBalance   MoneyModel         `bson:"closing_balance"`
	Status           string             `bson:"status"`
	TransactionUUIDs []string           `bson:"transaction_uuids"`
	CreatedAt        time.Time          `bson:"created_at"`
	UpdatedAt        time.Time          `bson:"updated_at"`
}

type MoneyModel struct {
	Amount   float64 `bson:"amount"`
	Currency string  `bson:"currency"`
}

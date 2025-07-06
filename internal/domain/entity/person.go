package entity

import (
	"time"

	"github.com/google/uuid"
)

type Person struct {
	ID        uuid.UUID
	Name      string
	Email     string
	Phone     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewPerson(name, email, phone string) *Person {
	now := time.Now()
	return &Person{
		ID:        uuid.New(),
		Name:      name,
		Email:     email,
		Phone:     phone,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func (p *Person) Update(name, email, phone string) {
	p.Name = name
	p.Email = email
	p.Phone = phone
	p.UpdatedAt = time.Now()
}
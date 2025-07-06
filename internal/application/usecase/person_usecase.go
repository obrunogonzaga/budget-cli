package usecase

import (
	"context"
	"fmt"

	"financli/internal/domain/entity"
	"financli/internal/domain/repository"
	"github.com/google/uuid"
)

type PersonUseCase struct {
	personRepo repository.PersonRepository
}

func NewPersonUseCase(personRepo repository.PersonRepository) *PersonUseCase {
	return &PersonUseCase{
		personRepo: personRepo,
	}
}

func (uc *PersonUseCase) CreatePerson(ctx context.Context, name, email, phone string) (*entity.Person, error) {
	person := entity.NewPerson(name, email, phone)
	
	if err := uc.personRepo.Create(ctx, person); err != nil {
		return nil, fmt.Errorf("failed to create person: %w", err)
	}
	
	return person, nil
}

func (uc *PersonUseCase) GetPerson(ctx context.Context, id uuid.UUID) (*entity.Person, error) {
	return uc.personRepo.FindByID(ctx, id)
}

func (uc *PersonUseCase) ListPeople(ctx context.Context) ([]*entity.Person, error) {
	return uc.personRepo.FindAll(ctx)
}

func (uc *PersonUseCase) UpdatePerson(ctx context.Context, id uuid.UUID, name, email, phone string) error {
	person, err := uc.personRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	
	person.Update(name, email, phone)
	
	return uc.personRepo.Update(ctx, person)
}

func (uc *PersonUseCase) DeletePerson(ctx context.Context, id uuid.UUID) error {
	return uc.personRepo.Delete(ctx, id)
}

func (uc *PersonUseCase) FindByEmail(ctx context.Context, email string) (*entity.Person, error) {
	return uc.personRepo.FindByEmail(ctx, email)
}
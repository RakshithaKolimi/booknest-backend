package publisher_service

import (
	"context"

	"github.com/google/uuid"

	"booknest/internal/domain"
)

type publisherService struct {
	r domain.PublisherRepository
}

func NewPublisherService(r domain.PublisherRepository) domain.PublisherService {
	return &publisherService{
		r: r,
	}
}

func (s *publisherService) FindByID(
	ctx context.Context,
	id uuid.UUID,
) (*domain.Publisher, error) {

	publisher, err := s.r.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &publisher, nil
}

func (s *publisherService) List(
	ctx context.Context,
	limit, offset int,
) ([]domain.Publisher, error) {
	return s.r.List(ctx, limit, offset)
}

func (s *publisherService) Create(
	ctx context.Context,
	input domain.PublisherInput,
) (*domain.Publisher, error) {

	publisher := &domain.Publisher{
		ID:          uuid.New(),
		LegalName:   input.LegalName,
		TradingName: input.TradingName,
		Email:       input.Email,
		Mobile:      input.Mobile,
		Address:     input.Address,
		City:        input.City,
		State:       input.State,
		Country:     input.Country,
		Zipcode:     input.Zipcode,
		IsActive:    true, // default active
	}

	if err := s.r.Create(ctx, publisher); err != nil {
		return nil, err
	}

	return publisher, nil
}

func (s *publisherService) Update(
	ctx context.Context,
	id uuid.UUID,
	input domain.PublisherInput,
) (*domain.Publisher, error) {

	publisher, err := s.r.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	publisher.LegalName = input.LegalName
	publisher.TradingName = input.TradingName
	publisher.Email = input.Email
	publisher.Mobile = input.Mobile
	publisher.Address = input.Address
	publisher.City = input.City
	publisher.State = input.State
	publisher.Country = input.Country
	publisher.Zipcode = input.Zipcode

	if err := s.r.Update(ctx, &publisher); err != nil {
		return nil, err
	}

	return &publisher, nil
}

func (s *publisherService) SetActive(
	ctx context.Context,
	id uuid.UUID,
	active bool,
) error {

	return s.r.SetActive(ctx, id, active)
}

func (s *publisherService) Delete(
	ctx context.Context,
	id uuid.UUID,
) error {

	return s.r.Delete(ctx, id)
}

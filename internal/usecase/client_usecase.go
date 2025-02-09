package usecase

import (
	"boonkosang/internal/repositories"
	"boonkosang/internal/requests"
	"boonkosang/internal/responses"
	"context"
	"errors"

	"github.com/google/uuid"
)

type ClientUsecase interface {
	Create(ctx context.Context, req requests.CreateClientRequest) (*responses.ClientResponse, error)
	Update(ctx context.Context, id uuid.UUID, req requests.UpdateClientRequest) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByID(ctx context.Context, id uuid.UUID) (*responses.ClientResponse, error)
	List(ctx context.Context, page, pageSize int) (*responses.ClientListResponse, error)
}

type clientUsecase struct {
	clientRepo repositories.ClientRepository
}

func NewClientUsecase(clientRepo repositories.ClientRepository) ClientUsecase {
	return &clientUsecase{
		clientRepo: clientRepo,
	}
}

func (u *clientUsecase) Create(ctx context.Context, req requests.CreateClientRequest) (*responses.ClientResponse, error) {
	existing, err := u.clientRepo.GetByEmail(ctx, req.Email)
	if err == nil && existing != nil {
		return nil, errors.New("client with this email already exists")
	}

	client, err := u.clientRepo.Create(ctx, req)
	if err != nil {
		return nil, err
	}

	return &responses.ClientResponse{
		ID:      client.ClientID,
		Name:    client.Name,
		Email:   client.Email,
		Tel:     client.Tel,
		Address: client.Address,
		TaxID:   client.TaxID,
	}, nil
}

func (u *clientUsecase) Update(ctx context.Context, id uuid.UUID, req requests.UpdateClientRequest) error {
	existing, err := u.clientRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if existing.Email != req.Email {
		client, err := u.clientRepo.GetByEmail(ctx, req.Email)
		if err == nil && client != nil {
			return errors.New("client with this email already exists")
		}
	}

	return u.clientRepo.Update(ctx, id, req)
}

func (u *clientUsecase) Delete(ctx context.Context, id uuid.UUID) error {
	return u.clientRepo.Delete(ctx, id)
}

func (u *clientUsecase) GetByID(ctx context.Context, id uuid.UUID) (*responses.ClientResponse, error) {
	client, err := u.clientRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &responses.ClientResponse{
		ID:      client.ClientID,
		Name:    client.Name,
		Email:   client.Email,
		Tel:     client.Tel,
		Address: client.Address,
		TaxID:   client.TaxID,
	}, nil
}

func (u *clientUsecase) List(ctx context.Context, page, pageSize int) (*responses.ClientListResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize
	clients, total, err := u.clientRepo.List(ctx, pageSize, offset)
	if err != nil {
		return nil, err
	}

	clientResponses := make([]responses.ClientResponse, len(clients))
	for i, client := range clients {
		clientResponses[i] = responses.ClientResponse{
			ID:      client.ClientID,
			Name:    client.Name,
			Email:   client.Email,
			Tel:     client.Tel,
			Address: client.Address,
			TaxID:   client.TaxID,
		}
	}

	return &responses.ClientListResponse{
		Clients: clientResponses,
		Total:   total,
	}, nil
}

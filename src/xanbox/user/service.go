package user

import (
	"context"

	"github.com/google/uuid"
)

type UserService interface {
	Create(ctx context.Context, user *User) error
	Get(ctx context.Context, id uuid.UUID) (*User, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type userService struct {
	repo UserRepository
}

func NewUserService(repo UserRepository) UserService {
	return &userService{
		repo: repo,
	}
}

func (s *userService) Create(
	ctx context.Context,
	user *User,
) error {
	err := s.repo.Create(ctx, user)

	return err
}

func (s *userService) Get(
	ctx context.Context,
	id uuid.UUID,
) (*User, error) {
	user, err := s.repo.Get(ctx, id)
	return user, err
}

func (s *userService) Delete(
	ctx context.Context,
	id uuid.UUID,
) error {
	return s.repo.Delete(ctx, id)
}

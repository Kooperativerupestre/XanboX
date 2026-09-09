package user

import (
	"context"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type User struct {
	ID   uuid.UUID `bun:"id, pk"`
	Name string    `bun:"name"`
}

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	Get(ctx context.Context, id uuid.UUID) (*User, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type userRepository struct {
	db *bun.DB
}

func NewUserRepository(db *bun.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(
	ctx context.Context,
	user *User,
) error {
	_, err := r.db.NewInsert().
		Model(user).
		Exec(ctx)

	return err
}

func (r *userRepository) Get(
	ctx context.Context,
	id uuid.UUID,
) (*User, error) {
	user := new(User)

	err := r.db.NewSelect().
		Model(user).
		Where("id = ?", id).
		Scan(ctx)

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *userRepository) Delete(
	ctx context.Context,
	id uuid.UUID,
) error {
	_, err := r.db.NewDelete().
		Model((*User)(nil)).
		Where("id = ?", id).
		Exec(ctx)

	return err
}

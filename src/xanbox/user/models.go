package user

import "github.com/google/uuid"

type User struct {
	ID   uuid.UUID `bun:"id, pk"`
	Name string    `bun:"name"`
}

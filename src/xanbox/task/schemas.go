package task

import "github.com/google/uuid"

type CreateTaskRequest struct {
	Language string
	Source   string
	Maker    uuid.UUID
}

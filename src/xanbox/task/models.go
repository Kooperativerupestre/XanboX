package task

import (
	"time"

	"github.com/google/uuid"
)

type Task struct {
	ID       uuid.UUID
	MadeAt   time.Time
	Language string
	Source   string
	Maker    uuid.UUID
}

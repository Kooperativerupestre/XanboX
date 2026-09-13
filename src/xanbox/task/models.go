package task

import (
	"time"

	"github.com/google/uuid"
)

type TaskStatus string

const (
	TaskPending    TaskStatus = "pending"
	TaskFailed     TaskStatus = "failed"
	TaskSuccessful TaskStatus = "successful"
)

func ValidateStatus(status string) bool {
	if status != "pending" && status != "failed" && status != "successful" {
		return false
	}
	return true
}

type SourceFile struct {
	Path string
	Code string
}

type Task struct {
	ID                     uuid.UUID    `bun:"id,pk"`
	MadeAt                 time.Time    `bun:"made_at"`
	Image                  string       `bun:"image"`
	EnvironmentPrepareCode []string     `bun:"environment_prepare_code"`
	ExecutionCode          string       `bun:"execution_code"`
	Source                 []SourceFile `bun:"source"`
	Maker                  uuid.UUID    `bun:"maker"`
	Status                 TaskStatus   `bun:"status"`
}

package task

import (
	"time"

	"github.com/Kooperativerupestre/XanboX/src/xanbox/task/source"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
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

type Task struct {
	ID                     uuid.UUID     `bun:"id,pk"`
	MadeAt                 time.Time     `bun:"made_at"`
	Image                  string        `bun:"image"`
	EnvironmentPrepareCode []string      `bun:"environment_prepare_code"`
	ExecutionCode          string        `bun:"execution_code"`
	Source                 []source.File `bun:"source"`
	Maker                  uuid.UUID     `bun:"maker"`
	Status                 TaskStatus    `bun:"status"`

	Execution *string `bun:"rel:has-one,join:id=task_id"`
}

type executionRecord struct {
	bun.BaseModel `bun:"table:executions"`

	ID string `bun:"id,pk"`
}

type taskExecutionRecord struct {
	bun.BaseModel `bun:"table:task_executions"`

	TaskID      uuid.UUID `bun:"task_id,pk"`
	ExecutionID string    `bun:"execution_id"`
}

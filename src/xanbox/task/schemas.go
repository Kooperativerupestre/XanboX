package task

import "github.com/google/uuid"

type CreateTaskRequest struct {
	Image                  string       `json:"image"`
	EnvironmentPrepareCode []string     `json:"environment_prepare_code"`
	Source                 []SourceFile `json:"source"`
	ExecutionCode          string       `json:"execution_code"`
	Maker                  uuid.UUID    `json:"maker"`
}

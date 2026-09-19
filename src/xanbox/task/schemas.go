package task

import (
	"github.com/Kooperativerupestre/XanboX/src/xanbox/task/source"
	"github.com/google/uuid"
)

type CreateTaskRequest struct {
	Image                  string        `json:"image"`
	EnvironmentPrepareCode []string      `json:"environment_prepare_code"`
	Source                 []source.File `json:"source"`
	ExecutionCode          string        `json:"execution_code"`
	Maker                  uuid.UUID     `json:"maker"`
}

type SyncTaskRequest struct {
	DockerID string `json:"docker_id"`
}

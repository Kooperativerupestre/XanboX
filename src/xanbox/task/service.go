package task

import (
	"context"

	"github.com/Kooperativerupestre/XanboX/src/xanbox/task/execution"
	"github.com/google/uuid"
)

type TaskService interface {
	Create(ctx context.Context, task *Task) (db_id uuid.UUID, docker_id string, err error, delete_err error)
	Get(ctx context.Context, id uuid.UUID) (*Task, error)
	Sync(ctx context.Context, dbID uuid.UUID) error
}

type taskService struct {
	repo TaskRepository
	tm   execution.TaskManager
}

func NewTaskService(repo TaskRepository, tm execution.TaskManager) TaskService {
	return &taskService{
		repo: repo,
		tm:   tm,
	}
}

func (s *taskService) Create(
	ctx context.Context,
	task *Task,
) (dbID uuid.UUID, dockerID string, err error, deleteErr error) {

	dockerID, err = s.tm.Create(ctx, task.Image, task.EnvironmentPrepareCode, task.Source, task.ExecutionCode)
	if err != nil {
		return uuid.Nil, "", err, nil
	}

	dbID, err = s.repo.Create(ctx, task, dockerID)
	if err != nil {
		_, deleteErr = s.tm.TryDelete(ctx, dockerID)

		return uuid.Nil, dockerID, err, deleteErr
	}

	return dbID, dockerID, nil, nil
}

func (s *taskService) Get(
	ctx context.Context,
	id uuid.UUID,
) (*Task, error) {
	task, err := s.repo.Get(ctx, id)
	return task, err
}

func (s *taskService) Sync(ctx context.Context, dbID uuid.UUID) error {
	task, err := s.repo.Get(ctx, dbID)
	if err != nil {
		return err
	}

	if task.Status != TaskPending {
		return nil
	}

	containerID, err := s.repo.GetContainerID(ctx, dbID)
	if err != nil {
		return err
	}

	isFinished, err := s.tm.IsFinished(ctx, containerID)
	if err != nil {
		return err
	}

	if !isFinished {
		return nil
	}

	isSuccessful, err := s.tm.Successful(ctx, containerID)
	if err != nil {
		return err
	}

	newStatus := TaskFailed
	if isSuccessful {
		newStatus = TaskSuccessful
	}

	updated, err := s.repo.UpdateStatus(ctx, dbID, newStatus)

	if err != nil {
		return err
	}

	if !updated {
		return nil
	}

	if _, err := s.tm.TryDelete(ctx, containerID); err != nil {
		return err
	}

	return s.repo.DeleteExecution(ctx, containerID)
}

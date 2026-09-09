package task

import (
	"context"

	"github.com/google/uuid"
)

type TaskService interface {
	Create(ctx context.Context, task *Task) error
	Get(ctx context.Context, id uuid.UUID) (*Task, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type taskService struct {
	repo TaskRepository
}

func NewTaskService(repo TaskRepository) TaskService {
	return &taskService{
		repo: repo,
	}
}

func (s *taskService) Create(
	ctx context.Context,
	task *Task,
) error {
	err := s.repo.Create(ctx, task)

	return err
}

func (s *taskService) Get(
	ctx context.Context,
	id uuid.UUID,
) (*Task, error) {
	task, err := s.repo.Get(ctx, id)
	return task, err
}

func (s *taskService) Delete(
	ctx context.Context,
	id uuid.UUID,
) error {
	return s.repo.Delete(ctx, id)
}

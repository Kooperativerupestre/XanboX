package task

import (
	"context"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type TaskRepository interface {
	Create(ctx context.Context, task *Task) error
	Get(ctx context.Context, id uuid.UUID) (*Task, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type taskRepository struct {
	db *bun.DB
}

func NewTaskRepository(db *bun.DB) TaskRepository {
	return &taskRepository{db: db}
}

func (r *taskRepository) Create(
	ctx context.Context,
	task *Task,
) error {
	_, err := r.db.NewInsert().Model(task).Exec(ctx)
	return err
}

func (r *taskRepository) Get(
	ctx context.Context,
	id uuid.UUID,
) (*Task, error) {
	task := new(Task)

	err := r.db.NewSelect().Model(task).Where("id = ?", id).Scan(ctx)

	if err != nil {
		return nil, err
	}

	return task, nil
}

func (r *taskRepository) Delete(
	ctx context.Context,
	id uuid.UUID,
) error {
	_, err := r.db.NewDelete().Model((*Task)(nil)).Where("id = ?", id).Exec(ctx)
	return err
}

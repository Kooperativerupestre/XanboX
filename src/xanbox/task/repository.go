package task

import (
	"context"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type TaskRepository interface {
	Create(ctx context.Context, task *Task, execID string) (id uuid.UUID, err error)
	Get(ctx context.Context, id uuid.UUID) (*Task, error)
	Delete(ctx context.Context, id uuid.UUID) error
	UpdateStatus(ctx context.Context, id uuid.UUID, newStatus TaskStatus) error
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
	containerID string,
) (uuid.UUID, error) {
	err := r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		if _, err := tx.NewInsert().
			Model(task).
			Returning("id").
			Exec(ctx); err != nil {
			return err
		}

		execution := &executionRecord{ID: containerID}
		if _, err := tx.NewInsert().Model(execution).Exec(ctx); err != nil {
			return err
		}

		link := &taskExecutionRecord{
			TaskID:      task.ID,
			ExecutionID: containerID,
		}
		_, err := tx.NewInsert().Model(link).Exec(ctx)

		return err
	})

	return task.ID, err
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

func (r *taskRepository) UpdateStatus(
	ctx context.Context,
	id uuid.UUID,
	newStatus TaskStatus,
) error {
	_, err := r.db.NewUpdate().
		Model((*Task)(nil)).
		Set("status = ?", newStatus).
		Where("id = ?", id).
		Exec(ctx)

	return err
}

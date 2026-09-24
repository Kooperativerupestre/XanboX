package execution

import (
	"context"
	"fmt"
	"time"

	"github.com/Kooperativerupestre/XanboX/src/xanbox/task/source"
	"github.com/moby/moby/client"
)

type TaskManager interface {
	IsFinished(ctx context.Context, id string) (bool, error)
	Failed(ctx context.Context, id string) (bool, error)
	Successful(ctx context.Context, id string) (bool, error)
	GetOutput(ctx context.Context, id string) (*ExecutionOutput, error)

	Create(
		ctx context.Context,
		image string,
		environmentPrepareCode []string,
		sourceFiles []source.File,
		executionCode string,
	) (id string, err error)

	TryDelete(ctx context.Context, id string) (TryDeleteOutput, error)
	Stop(ctx context.Context, id string) error
}

type taskManager struct {
	storage      *TaskExecutionsStorage
	dockerClient *client.Client
}

func NewTaskManager(storage *TaskExecutionsStorage) (*taskManager, error) {
	dockerClient, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, err
	}

	return &taskManager{
		storage:      storage,
		dockerClient: dockerClient,
	}, nil
}

func (tm *taskManager) IsFinished(
	ctx context.Context,
	id string,
) (bool, error) {
	execution, exists := tm.storage.get(id)
	if !exists {
		return false, fmt.Errorf("execution %q not found", id)
	}

	inspection, err := tm.dockerClient.ExecInspect(
		ctx,
		execution.ExecID(),
		client.ExecInspectOptions{},
	)
	if err != nil {
		return false, err
	}

	return !inspection.Running, nil
}

func (tm *taskManager) Failed(
	ctx context.Context,
	id string,
) (bool, error) {
	execution, exists := tm.storage.get(id)
	if !exists {
		return false, fmt.Errorf("execution %q not found", id)
	}

	inspection, err := tm.dockerClient.ExecInspect(
		ctx,
		execution.ExecID(),
		client.ExecInspectOptions{},
	)
	if err != nil {
		return false, err
	}

	if inspection.Running {
		return false, nil
	}

	return inspection.ExitCode != 0, nil
}

func (tm *taskManager) Successful(
	ctx context.Context,
	id string,
) (bool, error) {
	execution, exists := tm.storage.get(id)
	if !exists {
		return false, fmt.Errorf("execution %q not found", id)
	}

	inspection, err := tm.dockerClient.ExecInspect(
		ctx,
		execution.ExecID(),
		client.ExecInspectOptions{},
	)
	if err != nil {
		return false, err
	}

	if inspection.Running {
		return false, nil
	}

	return inspection.ExitCode == 0, nil
}

func (tm *taskManager) GetOutput(
	ctx context.Context,
	id string,
) (*ExecutionOutput, error) {
	execution, exists := tm.storage.get(id)
	if !exists {
		return nil, fmt.Errorf("execution %q not found", id)
	}

	return &ExecutionOutput{
		Stdout: execution.Stdout.String(),
		Stderr: execution.Stderr.String(),
	}, nil
}

func (tm *taskManager) Create(
	ctx context.Context,
	image string,
	environmentPrepareCode []string,
	sourceFiles []source.File,
	executionCode string,
) (id string, err error) {
	execution, err := initialize(
		ctx,
		tm.dockerClient,
		image,
	)
	if err != nil {
		return "", err
	}

	needsCleanup := true

	defer func() {
		if !needsCleanup {
			return
		}

		_, _ = tm.dockerClient.ContainerRemove(
			context.Background(),
			execution.ContainerID,
			client.ContainerRemoveOptions{},
		)
	}()

	if err := resolveDependencies(
		ctx,
		execution,
		tm.dockerClient,
		environmentPrepareCode,
	); err != nil {
		return "", err
	}

	if err := addSourceFiles(
		ctx,
		execution,
		tm.dockerClient,
		sourceFiles,
	); err != nil {
		return "", err
	}

	tm.storage.add(execution)

	needsCleanup = false

	go func() {
		_ = execute(
			context.Background(),
			execution,
			tm.dockerClient,
			executionCode,
		)
	}()

	return execution.ExecID(), nil
}

func (tm *taskManager) Stop(
	ctx context.Context,
	id string,
) error {
	execution, exists := tm.storage.get(id)
	if !exists {
		return nil
	}

	inspection, err := tm.dockerClient.ContainerInspect(
		ctx,
		execution.ContainerID,
		client.ContainerInspectOptions{},
	)
	if err != nil {
		return err
	}

	if !inspection.Container.State.Running {
		return nil
	}

	const maxRetries = 5
	const retryInterval = time.Second

	for attempt := 0; attempt < maxRetries; attempt++ {
		_, err = tm.dockerClient.ContainerStop(
			ctx,
			execution.ContainerID,
			client.ContainerStopOptions{},
		)

		if err == nil {
			return nil
		}

		inspection, inspectErr := tm.dockerClient.ContainerInspect(
			ctx,
			execution.ContainerID,
			client.ContainerInspectOptions{},
		)

		if inspectErr == nil && !inspection.Container.State.Running {
			return nil
		}

		if attempt == maxRetries-1 {
			return err
		}

		timer := time.NewTimer(retryInterval)

		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()

		case <-timer.C:
		}
	}

	return nil
}

type TryDeleteOutput uint8

const (
	TryDeleteDeleted TryDeleteOutput = iota
	TryDeleteRunning
)

func (tm *taskManager) TryDelete(
	ctx context.Context,
	id string,
) (TryDeleteOutput, error) {
	execution, exists := tm.storage.get(id)
	if !exists {
		return TryDeleteDeleted, fmt.Errorf(
			"execution %q not found",
			id,
		)
	}

	inspection, err := tm.dockerClient.ExecInspect(
		ctx,
		execution.ExecID(),
		client.ExecInspectOptions{},
	)
	if err != nil {
		return TryDeleteDeleted, err
	}

	if inspection.Running {
		return TryDeleteRunning, nil
	}

	_, err = tm.dockerClient.ContainerRemove(
		ctx,
		execution.ContainerID,
		client.ContainerRemoveOptions{},
	)
	if err != nil {
		return TryDeleteDeleted, err
	}

	tm.storage.delete(id)

	return TryDeleteDeleted, nil
}

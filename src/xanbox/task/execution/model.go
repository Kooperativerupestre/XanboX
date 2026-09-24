package execution

import (
	"context"
	"sync"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
)

type ExecutionOutput struct {
	Stdout string
	Stderr string
}

type executionContainer struct {
	mu          sync.RWMutex
	execID      string
	ContainerID string
	Stdout      synchronizedBuffer
	Stderr      synchronizedBuffer
}

func (exC *executionContainer) ExecID() string {
	exC.mu.RLock()
	defer exC.mu.RUnlock()
	return exC.execID
}

func (exC *executionContainer) AddExecID(newID string) {
	exC.mu.Lock()
	defer exC.mu.Unlock()
	exC.execID = newID
}

func NewExecutionContainer(containerID string) *executionContainer {
	return &executionContainer{
		ContainerID: containerID,
	}
}

func initialize(
	ctx context.Context,
	dockerClient *client.Client,
	image string,
) (*executionContainer, error) {
	response, err := dockerClient.ContainerCreate(
		ctx,
		client.ContainerCreateOptions{
			Config: &container.Config{
				Image:      image,
				WorkingDir: "/workspace",
			},
		},
	)
	if err != nil {
		return nil, err
	}

	_, err = dockerClient.ContainerStart(
		ctx,
		response.ID,
		client.ContainerStartOptions{},
	)
	if err != nil {
		return nil, err
	}

	return NewExecutionContainer(response.ID), nil
}

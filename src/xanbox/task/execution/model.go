package execution

import (
	"context"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
)

type ExecutionOutput struct {
	Stdout string
	Stderr string
}

type executionContainer struct {
	ExecID      string
	ContainerID string
	Stdout      synchronizedBuffer
	Stderr      synchronizedBuffer
}

func NewExecutionContainer(id string) *executionContainer {
	return &executionContainer{
		ExecID: id,
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

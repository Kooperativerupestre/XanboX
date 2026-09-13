package taskEnvironment

import (
	"context"
	"fmt"
	"io"
)

type ContainerExecutor struct {
}

func NewContainerExecutor() *ContainerExecutor {
	return &ContainerExecutor{}
}

func (ce *ContainerExecutor) Execute(ctx context.Context, c *Container, executionCode string) (string, error) {
	exitCode, output, err := c.Ctr.Exec(ctx, []string{"sh", "-c", executionCode})
	if err != nil {
		return "", err
	}

	outputBytes, err := io.ReadAll(output)
	if err != nil {
		return "", err
	}

	if exitCode != 0 {
		return string(outputBytes), fmt.Errorf("execution exited with code %d", exitCode)
	}

	return string(outputBytes), nil
}

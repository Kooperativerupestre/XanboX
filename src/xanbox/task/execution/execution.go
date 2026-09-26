package execution

import (
	"context"
	"fmt"

	"github.com/moby/moby/api/pkg/stdcopy"
	"github.com/moby/moby/client"
)

func createExecution(
	ctx context.Context,
	c *executionContainer,
	dockerClient *client.Client,
	executionCode string,
) error {
	exec, err := dockerClient.ExecCreate(
		ctx,
		c.ContainerID,
		client.ExecCreateOptions{
			Cmd: []string{"sh", "-c", executionCode},
		},
	)

	if err != nil {
		return err
	}
	c.AddExecID(exec.ID)
	return nil
}

func execute(
	ctx context.Context,
	c *executionContainer,
	dockerClient *client.Client,
	executionCode string,
) error {

	response, err := dockerClient.ExecAttach(
		ctx,
		c.ExecID(),
		client.ExecAttachOptions{},
	)
	if err != nil {
		return err
	}

	_, err = stdcopy.StdCopy(
		&c.Stdout,
		&c.Stderr,
		response.Reader,
	)
	response.Close()

	if err != nil {
		return err
	}

	inspection, err := dockerClient.ExecInspect(
		ctx,
		c.ExecID(),
		client.ExecInspectOptions{},
	)
	if err != nil {
		return err
	}

	if inspection.ExitCode != 0 {
		return fmt.Errorf(
			"execution failed with exit code %d",
			inspection.ExitCode,
		)
	}

	return nil
}

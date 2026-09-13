package taskEnvironment

import (
	"context"
	"fmt"
	"io"
	"path/filepath"

	"github.com/Kooperativerupestre/XanboX/src/xanbox/task"
	"github.com/docker/go-sdk/container"
)

type DockerPrepare struct {
}

func NewDockerPrepare() *DockerPrepare {
	return &DockerPrepare{}
}

type Container struct {
	Ctr *container.Container
}

func NewContainer(ctr *container.Container) *Container {
	return &Container{Ctr: ctr}
}

func (*DockerPrepare) Prepare(ctx context.Context, task *task.Task) (*Container, []string, error) {
	ctr, err := container.Run(ctx, container.WithImage(task.Image))

	var outputs []string
	if err != nil {
		return nil, outputs, err
	}

	cleanup := true
	defer func() {
		if cleanup {
			_ = ctr.Terminate(ctx)
		}
	}()

	_, _, err = ctr.Exec(ctx, []string{"mkdir", "-p", "/workspace"})
	if err != nil {
		return nil, outputs, err
	}
	for _, command := range task.EnvironmentPrepareCode {
		exec := []string{"sh", "-c", command}

		exitCode, output, err := ctr.Exec(ctx, exec)

		outputBytes, outputReadErr := io.ReadAll(output)
		if outputReadErr != nil {
			return nil, outputs, outputReadErr
		}

		outputs = append(outputs, string(outputBytes))

		if err != nil {
			return nil, outputs, err
		}
		if exitCode != 0 {
			return nil, outputs, fmt.Errorf("preparation command exited with code %d", exitCode)
		}
	}
	cleanup = false

	if err != nil {
		return nil, outputs, err
	}
	return NewContainer(ctr), outputs, nil
}

func (c *Container) WriteFile(ctx context.Context, source_file task.SourceFile) error {
	fullPath := filepath.Join("/workspace", source_file.Path)
	_, _, err := c.Ctr.Exec(ctx, []string{"mkdir", "-p", filepath.Dir(fullPath)})

	if err != nil {
		return err
	}
	data := []byte(source_file.Code)

	return c.Ctr.CopyToContainer(
		ctx,
		data,
		fullPath,
		0o644,
	)
}

func (c *Container) WriteAllFiles(ctx context.Context, t *task.Task) error {
	for _, sf := range t.Source {
		err := c.WriteFile(ctx, sf)
		if err != nil {
			return err
		}
	}
	return nil
}

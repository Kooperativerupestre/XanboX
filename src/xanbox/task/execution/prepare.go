package execution

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"

	"github.com/Kooperativerupestre/XanboX/src/xanbox/task/source"
	"github.com/moby/moby/api/pkg/stdcopy"
	"github.com/moby/moby/client"
)

func resolveDependencies(
	ctx context.Context,
	c *executionContainer,
	dockerClient *client.Client,
	commands []string,
) error {
	for _, command := range commands {
		exec, err := dockerClient.ExecCreate(
			ctx,
			c.execID,
			client.ExecCreateOptions{
				Cmd: []string{"sh", "-c", command},
			},
		)
		if err != nil {
			return err
		}

		response, err := dockerClient.ExecAttach(
			ctx,
			exec.ID,
			client.ExecAttachOptions{},
		)
		if err != nil {
			return err
		}
		defer response.Close()

		_, err = stdcopy.StdCopy(
			&c.Stdout,
			&c.Stderr,
			response.Reader,
		)
		if err != nil {
			return err
		}

		inspection, err := dockerClient.ExecInspect(
			ctx,
			exec.ID,
			client.ExecInspectOptions{},
		)
		if err != nil {
			return err
		}

		if inspection.ExitCode != 0 {
			return fmt.Errorf(
				"dependency command failed with exit code %d: %s",
				inspection.ExitCode,
				command,
			)
		}
	}

	return nil
}

func addSourceFiles(
	ctx context.Context,
	c *executionContainer,
	dockerClient *client.Client,
	source []source.File,
) error {
	var buffer bytes.Buffer
	tarWriter := tar.NewWriter(&buffer)

	for _, sourceFile := range source {
		header := &tar.Header{
			Name: sourceFile.Path,
			Mode: 0644,
			Size: int64(len(sourceFile.Code)),
		}

		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		if _, err := tarWriter.Write([]byte(sourceFile.Code)); err != nil {
			return err
		}
	}

	if err := tarWriter.Close(); err != nil {
		return err
	}

	_, err := dockerClient.CopyToContainer(
		ctx,
		c.execID,
		client.CopyToContainerOptions{
			DestinationPath: "/workspace",
			Content:         &buffer,
		},
	)
	if err != nil {
		return err
	}

	return nil
}

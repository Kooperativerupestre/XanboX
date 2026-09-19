package execution

import (
	"context"

	"github.com/moby/moby/client"
)

func inspectExecExitCode(ctx context.Context, cli *client.Client, execID string) (int, error) {
	inspection, err := cli.ExecInspect(ctx, execID, client.ExecInspectOptions{})

	if err != nil {
		return 0, err
	}
	return inspection.ExitCode, nil
}

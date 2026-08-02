package runner

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"
)

type CommandRunner interface {
	Run(ctx context.Context, stdin io.Reader, name string, args ...string) (Result, error)
}

type Result struct {
	Stdout string
	Stderr string
}

type OSCommandRunner struct{}

func (OSCommandRunner) Run(ctx context.Context, stdin io.Reader, name string, args ...string) (Result, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdin = stdin

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	result := Result{Stdout: stdout.String(), Stderr: stderr.String()}
	if err != nil {
		return result, fmt.Errorf("%s %v failed: %w", name, args, err)
	}
	return result, nil
}

package internal

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
)

var setupSucceeded = false //nolint:gochecknoglobals

func setup(ctx context.Context) error {
	if setupSucceeded {
		return nil
	}
	cmd := exec.CommandContext(ctx, "golangci-lint", "version")
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("error running command: %w", err)
	}
	setupSucceeded = true
	return nil
}

// spell-checker: ignore golangci

func Lint(ctx context.Context, path string) (bool, error) {
	err := setup(ctx)
	if err != nil {
		return false, fmt.Errorf("error checking golangci-lint installation: %w", err)
	}
	cmd := exec.CommandContext(ctx, "golangci-lint", "run")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = path
	err = cmd.Run()
	if err != nil {
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			return false, nil
		}
		return false, fmt.Errorf("error running command: %w", err)
	}

	return true, nil
}

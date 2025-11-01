package internal

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
)

type Tool interface {
	Run(ctx context.Context, path string) (bool, error)
}

type commandLineTool struct {
	command   string
	checked   bool
	checkArgs []string
	mainArgs  []string
}

func NewCommandLineTool(command string, checkArgs []string, mainArgs []string) Tool {
	return &commandLineTool{
		command:   command,
		checked:   len(checkArgs) == 0,
		checkArgs: checkArgs,
		mainArgs:  mainArgs,
	}
}

type commandLineToolError struct {
	tool  *commandLineTool
	inner error
}

func (e commandLineToolError) Error() string {
	return fmt.Sprintf("error running command %s: %s", e.tool.command, e.inner)
}

func (e commandLineToolError) Unwrap() error {
	return e.inner
}

// spell-checker: ignore gosec

func (tool *commandLineTool) Run(ctx context.Context, path string) (bool, error) {
	err := tool.setup(ctx)
	if err != nil {
		return false, err
	}
	cmd := exec.CommandContext(ctx, tool.command, tool.mainArgs...) //nolint:gosec
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = path
	err = cmd.Run()
	if err != nil {
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			return false, nil
		}
		return false, commandLineToolError{tool: tool, inner: err}
	}

	return true, nil
}

func (tool *commandLineTool) setup(ctx context.Context) error {
	if tool.checked {
		return nil
	}
	cmd := exec.CommandContext(ctx, tool.command, tool.checkArgs...) //nolint:gosec
	err := cmd.Run()
	if err != nil {
		return commandLineToolError{tool: tool, inner: err}
	}
	tool.checked = true
	return nil
}

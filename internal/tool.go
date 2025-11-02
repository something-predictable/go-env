package internal

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
)

type Tool interface {
	Run(ctx context.Context, path string, files []string) (bool, error)
}

type commandLineTool struct {
	command   string
	checked   bool
	checkArgs []string
	mainArgs  func(files []string) []string
	result    func(stdout string, exitCode int, diagnostics io.Writer) (bool, error)
}

func Filter(files []string, predicate func(file string) bool) []string {
	passed := make([]string, 0, len(files))
	for _, f := range files {
		if predicate(f) {
			passed = append(passed, f)
		}
	}
	return passed
}

func NewCommandLineTool(
	command string,
	checkArgs []string,
	mainArgs func(files []string) []string,
	result func(stdout string, exitCode int, diagnostics io.Writer) (bool, error),
) Tool {
	return &commandLineTool{
		command:   command,
		checked:   len(checkArgs) == 0,
		checkArgs: checkArgs,
		mainArgs:  mainArgs,
		result:    result,
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

func (tool *commandLineTool) Run(ctx context.Context, path string, files []string) (bool, error) {
	err := tool.setup(ctx)
	if err != nil {
		return false, err
	}
	args := tool.mainArgs(files)
	if args == nil {
		return true, nil
	}
	var stdout bytes.Buffer
	cmd := exec.CommandContext(ctx, tool.command, tool.mainArgs(files)...) //nolint:gosec
	cmd.Stdout = &stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = path
	err = cmd.Run()
	if err != nil {
		var exitError *exec.ExitError
		if !errors.As(err, &exitError) {
			return false, commandLineToolError{tool: tool, inner: err}
		}
		if tool.result != nil {
			return tool.result(stdout.String(), exitError.ExitCode(), os.Stdout)
		}
		_, err := fmt.Fprint(os.Stdout, stdout.String())
		if err != nil {
			return false, err //nolint:wrapcheck
		}
		return false, nil
	}

	if tool.result != nil {
		return tool.result(stdout.String(), 0, os.Stdout)
	}
	return true, nil
}

func (tool *commandLineTool) setup(ctx context.Context) error {
	if tool.checked {
		return nil
	}
	cmd := exec.CommandContext(ctx, tool.command, tool.checkArgs...) //nolint:gosec
	err := cmd.Run()
	if ctx.Err() != nil {
		return ctx.Err() //nolint:wrapcheck
	}
	if err != nil {
		return commandLineToolError{tool: tool, inner: err}
	}
	tool.checked = true
	return nil
}

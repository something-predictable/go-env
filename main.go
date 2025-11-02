package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"riddance/env/internal"
	"slices"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"
)

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigCh
		cancel()
	}()

	err := run(ctx)
	if err != nil {
		log.Fatal(err)
	}
}

// spell-checker: ignore forbidigo

func run(ctx context.Context) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting working directory: %w", err)
	}
	err = internal.WatchSource(ctx, cwd,
		func(ctx context.Context, files []string) error {
			return runChecks(ctx, cwd, files)
		},
		func(ctx context.Context, files []string, deleted bool) error {
			now := time.Now().Format(time.TimeOnly)
			for _, path := range files {
				if deleted {
					fmt.Printf("🗑️ %s - %s deleted\n", now, path) //nolint:forbidigo
				} else {
					fmt.Printf("💾 %s - %s saved\n", now, path) //nolint:forbidigo
				}
			}
			err := runChecks(ctx, cwd, files)
			if err != nil {
				return fmt.Errorf("error running checks: %w", err)
			}
			return nil
		})
	if err != nil {
		return fmt.Errorf("error watching files: %w", err)
	}

	return nil
}

func runChecks(ctx context.Context, path string, files []string) error {
	success, err := check(ctx, path, files)
	if err != nil {
		return fmt.Errorf("error checking project: %w", err)
	}
	if success {
		fmt.Println("🚀  All good 👌") //nolint:forbidigo
	} else {
		fmt.Println("⚠️  Issues found 👆") //nolint:forbidigo
	}
	return nil
}

func check(ctx context.Context, path string, files []string) (bool, error) {
	tools := []internal.Tool{
		internal.LinterTool(),
		internal.SpellCheckerTool(),
		internal.TestTool(),
	}

	success := make([]bool, len(tools))
	// spell-checker: ignore errgroup
	group := errgroup.Group{}
	for ix, tool := range tools {
		group.Go(func() error {
			s, err := tool.Run(ctx, path, files)
			success[ix] = s
			return err //nolint:wrapcheck
		})
	}
	err := group.Wait()
	if err != nil {
		return false, fmt.Errorf("error running checker tool: %w", err)
	}

	return !slices.Contains(success, false), nil
}

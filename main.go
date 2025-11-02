package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"riddance/env/internal"
	"syscall"
	"time"
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
	lintSuccess, err := internal.LinterTool().Run(ctx, path, files)
	if err != nil {
		return false, fmt.Errorf("linter error: %w", err)
	}
	spellSuccess, err := internal.SpellCheckerTool().Run(ctx, path, files)
	if err != nil {
		return false, fmt.Errorf("spell checker error: %w", err)
	}
	testSuccess, err := internal.TestTool().Run(ctx, path, files)
	if err != nil {
		return false, fmt.Errorf("test error: %w", err)
	}
	return lintSuccess && spellSuccess && testSuccess, nil
}

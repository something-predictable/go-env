package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"riddance/env/internal"
	"time"
)

func main() {
	err := run()
	if err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting working directory: %w", err)
	}
	err = internal.WatchSource(cwd,
		func() error {
			return runChecks(cwd)
		},
		func(path string, deleted bool) error {
			rel, err := filepath.Rel(cwd, path)
			if err != nil {
				return fmt.Errorf("error getting working directory: %w", err)
			}
			now := time.Now().Format(time.TimeOnly)
			if deleted {
				fmt.Printf("🗑️ %s - %s deleted\n", now, rel) //nolint:forbidigo
			} else {
				fmt.Printf("💾 %s - %s saved\n", now, rel) //nolint:forbidigo
			}
			err = runChecks(cwd)
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

func runChecks(path string) error {
	ctx := context.Background()
	success, err := check(ctx, path)
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

func check(ctx context.Context, path string) (bool, error) {
	success, err := internal.Lint(ctx, path)
	if err != nil {
		return false, fmt.Errorf("lint errors: %w", err)
	}
	return success, nil
}

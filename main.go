package main

import (
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
	err = internal.WatchSource(cwd, func(path string, deleted bool) error {
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
		return nil
	})
	if err != nil {
		return fmt.Errorf("error watching files: %w", err)
	}

	return nil
}

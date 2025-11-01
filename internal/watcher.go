package internal

// spell-checker: ignore fsnotify

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/fsnotify/fsnotify"
)

type WatcherEventError struct {
	event fsnotify.Event
}

func (e WatcherEventError) Error() string {
	return fmt.Sprintf("error event from filesystem: %s", e.event)
}

func WatchSource(
	ctx context.Context,
	path string,
	init func() error,
	onChange func(path string, removed bool) error,
) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("error creating watcher: %w", err)
	}

	defer func() {
		err := watcher.Close()
		if err != nil {
			log.Fatalf("Error closing watcher: %v", err) //nolint: revive
		}
	}()

	err = watcher.Add(path)
	if err != nil {
		return fmt.Errorf("error adding watcher directory: %w", err)
	}

	err = init()
	if err != nil {
		return fmt.Errorf("error performing initialization before watcher starts: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case event, ok := <-watcher.Events:
			if !ok {
				return WatcherEventError{event: event}
			}
			if !strings.HasSuffix(event.Name, ".go") && !strings.HasSuffix(event.Name, "go.mod") {
				continue
			}
			err := onChange(event.Name, event.Op == fsnotify.Remove)
			if err != nil {
				return fmt.Errorf("error handling file change: %w", err)
			}
		case err := <-watcher.Errors:
			return fmt.Errorf("error from filesystem watcher: %w", err)
		}
	}
}

package internal

// spell-checker: ignore fsnotify

import (
	"context"
	"errors"
	"fmt"
	"log"
	"path/filepath"
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
	init func(ctx context.Context) error,
	onChange func(ctx context.Context, paths []string, removed bool) error,
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

	err = init(ctx)
	if err != nil {
		return fmt.Errorf("error performing initialization before watcher starts: %w", err)
	}

	handlerErrors := make(chan error, 1)
	defer close(handlerErrors)
	filesHandled := make(chan fileQueueSnapshot, 1)
	defer close(filesHandled)

	queue := fileQueue{
		clock:    0,
		fileAges: make(map[string]uint64),
	}

	cancel := func() {}
	for {
		select {
		case <-ctx.Done():
			cancel()
			return nil
		case event, ok := <-watcher.Events:
			cancel()
			if !ok {
				return WatcherEventError{event: event}
			}
			if !isSource(event.Name) {
				continue
			}
			// spell-checker: ignore fatcontext
			changeCtx, changeCancel := context.WithCancel(ctx) //nolint:fatcontext
			cancel = changeCancel
			rel, err := filepath.Rel(path, event.Name)
			if err != nil {
				return fmt.Errorf("error getting relative path: %w", err)
			}
			queue.addFile(rel)
			modified := queue.snapshot()
			go func() {
				err := onChange(changeCtx, modified.files, event.Op == fsnotify.Remove)
				if err != nil {
					if errors.Is(err, context.Canceled) {
						return
					}
					cancel()
					handlerErrors <- fmt.Errorf("error handling file change: %w", err)
					return
				}
				filesHandled <- modified
			}()
		case files := <-filesHandled:
			queue.handled(&files)
		case err := <-handlerErrors:
			cancel()
			return fmt.Errorf("error from filesystem event handler: %w", err)
		case err := <-watcher.Errors:
			cancel()
			return fmt.Errorf("error from filesystem watcher: %w", err)
		}
	}
}

func isSource(name string) bool {
	return strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "go.mod")
}

type fileQueue struct {
	clock    uint64
	fileAges map[string]uint64
}

func (q *fileQueue) addFile(name string) {
	q.fileAges[name] = q.clock
	q.clock++
}

type fileQueueSnapshot struct {
	clock uint64
	files []string
}

func (q *fileQueue) snapshot() fileQueueSnapshot {
	clock := q.clock
	files := make([]string, 0, len(q.fileAges))
	for file := range q.fileAges {
		files = append(files, file)
	}

	return fileQueueSnapshot{
		clock: clock,
		files: files,
	}
}

func (q *fileQueue) handled(snapshot *fileQueueSnapshot) {
	for _, file := range snapshot.files {
		age := q.fileAges[file]
		if age < snapshot.clock {
			delete(q.fileAges, file)
		}
	}
}

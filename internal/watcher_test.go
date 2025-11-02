package internal_test

// spell-checker: ignore errgroup

import (
	"context"
	"os"
	"path"
	"riddance/env/internal"
	"slices"
	"testing"
	"time"

	"golang.org/x/sync/errgroup"
)

func TestNoChange(t *testing.T) {
	t.Parallel()
	harness(t, func(_ string) error {
		return nil
	}, func(events []event) {
		assertEvents(t, events, []event{})
	})
}

func TestSingleFileCreate(t *testing.T) {
	t.Parallel()
	harness(t, func(dir string) error {
		err := writeFile(dir, "goat.go")
		if err != nil {
			return err
		}
		return nil
	}, func(events []event) {
		assertEvents(t, events, []event{
			{paths: []string{"goat.go"}, removed: false},
		})
	})
}

func TestMultiFileCreate(t *testing.T) {
	t.Parallel()
	harness(t, func(dir string) error {
		err := writeFile(dir, "goat1.go")
		if err != nil {
			return err
		}
		err = writeFile(dir, "goat1.go")
		if err != nil {
			return err
		}
		err = writeFile(dir, "goat2.go")
		if err != nil {
			return err
		}
		return err
	}, func(events []event) {
		assertEvents(t, events, []event{
			{paths: []string{"goat1.go", "goat2.go"}, removed: false},
		})
	})
}

func TestQuickMultiFileCreate(t *testing.T) {
	t.Parallel()
	harness(t, func(dir string) error {
		group := errgroup.Group{}
		group.Go(func() error {
			return writeFile(dir, "goat1.go")
		})
		group.Go(func() error {
			return writeFile(dir, "goat2.go")
		})
		group.Go(func() error {
			return writeFile(dir, "goat1.go")
		})
		return group.Wait()
	}, func(events []event) {
		assertEvents(t, events, []event{
			{paths: []string{"goat1.go", "goat2.go"}, removed: false},
		})
	})
}

func TestSingleFileDelete(t *testing.T) {
	t.Parallel()
	harness(t, func(dir string) error {
		err := writeFile(dir, "goat.go")
		if err != nil {
			return err
		}
		time.Sleep(100 * time.Millisecond)
		err = os.Remove(path.Join(dir, "goat.go"))
		if err != nil {
			return err
		}
		return nil
	}, func(events []event) {
		assertEvents(t, events, []event{
			{paths: []string{"goat.go"}, removed: false},
			{paths: []string{"goat.go"}, removed: true},
		})
	})
}

func writeFile(dir string, name string) error {
	return os.WriteFile(path.Join(dir, name), []byte{}, 0600)
}

type event = struct {
	paths   []string
	removed bool
}

func eventCompare(one event, other event) int {
	if one.removed != other.removed {
		return -1
	}
	return slices.Compare(
		slices.Sorted(slices.Values(one.paths)),
		slices.Sorted(slices.Values(other.paths)),
	)
}

func assertEvents(t *testing.T, actual []event, expected []event) {
	t.Helper()
	if slices.CompareFunc(actual, expected, eventCompare) != 0 {
		t.Errorf("Unexpected events:\n  got %v\n  expected %v", actual, expected)
	}
}

func harness(t *testing.T, actions func(dir string) error, assert func(events []event)) {
	t.Helper()
	dir := t.TempDir()

	const eventQueueSize = 4
	removedFiles := make(chan []string, eventQueueSize)
	defer close(removedFiles)
	changedFiles := make(chan []string, eventQueueSize)
	defer close(changedFiles)
	watcherErrors := make(chan error, 1)
	defer close(watcherErrors)
	actionErrors := make(chan error, 1)
	defer close(actionErrors)
	initialized := make(chan struct{}, 1)
	defer close(initialized)

	ctx, cancel := context.WithCancel(t.Context())
	go func() {
		err := internal.WatchSource(ctx, dir, func(_ context.Context) error {
			initialized <- struct{}{}
			return nil
		}, func(ctx context.Context, paths []string, removed bool) error {
			err := delay(ctx, 10*time.Millisecond)
			if err != nil {
				return err
			}
			if ctx.Err() != nil {
				return ctx.Err()
			}
			if removed {
				removedFiles <- paths
			} else {
				changedFiles <- paths
			}
			return nil
		})

		if err != nil {
			watcherErrors <- err
		}
	}()

	go func() {
		<-initialized
		err := actions(dir)
		if err != nil {
			actionErrors <- err
		}
		// Wait for file system to send events
		time.Sleep(500 * time.Millisecond)
		cancel()
	}()

	events := make([]event, 0)
loop:
	for {
		select {
		case paths := <-changedFiles:
			events = append(events, event{paths: paths, removed: false})
		case paths := <-removedFiles:
			events = append(events, event{paths: paths, removed: true})
		case err := <-watcherErrors:
			t.Fatalf("Watcher failed: %v", err)
		case err := <-actionErrors:
			t.Fatalf("Action error: %v", err)
		case <-ctx.Done():
			break loop
		}
	}

	assert(events)
}

func delay(ctx context.Context, d time.Duration) error {
	select {
	case <-time.After(d):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

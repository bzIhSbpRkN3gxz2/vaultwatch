package scheduler

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestScheduler_RunsJobOnTick(t *testing.T) {
	var count int32
	s := New(20*time.Millisecond, func(ctx context.Context) error {
		atomic.AddInt32(&count, 1)
		return nil
	})

	ctx, cancel := context.WithTimeout(context.Background(), 70*time.Millisecond)
	defer cancel()
	s.Run(ctx)

	if got := atomic.LoadInt32(&count); got < 2 {
		t.Errorf("expected at least 2 invocations, got %d", got)
	}
}

func TestScheduler_CallsOnError(t *testing.T) {
	var errCount int32
	s := New(20*time.Millisecond, func(ctx context.Context) error {
		return errors.New("boom")
	})
	s.OnError = func(err error) {
		atomic.AddInt32(&errCount, 1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 55*time.Millisecond)
	defer cancel()
	s.Run(ctx)

	if got := atomic.LoadInt32(&errCount); got < 1 {
		t.Errorf("expected OnError to be called, got %d", got)
	}
}

func TestScheduler_RunOnce_ExecutesImmediately(t *testing.T) {
	var count int32
	s := New(1*time.Hour, func(ctx context.Context) error {
		atomic.AddInt32(&count, 1)
		return nil
	})

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	s.RunOnce(ctx)

	if got := atomic.LoadInt32(&count); got != 1 {
		t.Errorf("expected 1 immediate invocation, got %d", got)
	}
}

func TestScheduler_StopsOnContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	done := make(chan struct{})
	s := New(10*time.Millisecond, func(ctx context.Context) error { return nil })
	go func() {
		s.Run(ctx)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(200 * time.Millisecond):
		t.Error("scheduler did not stop after context cancel")
	}
}

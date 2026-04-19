package scheduler

import (
	"context"
	"time"
)

// Job is a function executed on each tick.
type Job func(ctx context.Context) error

// Scheduler runs a job at a fixed interval until the context is cancelled.
type Scheduler struct {
	interval time.Duration
	job      Job
	OnError  func(err error)
}

// New creates a Scheduler with the given interval and job.
func New(interval time.Duration, job Job) *Scheduler {
	return &Scheduler{
		interval: interval,
		job:      job,
		OnError:  func(err error) {},
	}
}

// Run starts the scheduler loop, blocking until ctx is done.
func (s *Scheduler) Run(ctx context.Context) {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := s.job(ctx); err != nil {
	ce executes the job immediately starts the interval loop.
func (s *Scheduler) RunOnce(ctx context.Context) {
	if err := s.job(ctx); err != nil {
		s.OnError(err)
	}
	s.Run(ctx)
}

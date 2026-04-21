// Package pipeline provides a composable lease-processing pipeline that
// chains multiple processing stages (filter, sample, dedup, throttle, dispatch)
// into a single reusable unit.
package pipeline

import (
	"context"
	"fmt"

	"github.com/your-org/vaultwatch/internal/lease"
)

// Stage is a single processing step in the pipeline. It receives a lease and
// returns either the (possibly mutated) lease to continue processing, or an
// error. Returning a wrapped ErrSkip causes the pipeline to silently drop the
// lease without propagating an error to the caller.
type Stage interface {
	Process(ctx context.Context, l *lease.Lease) (*lease.Lease, error)
}

// StageFunc is a function adapter that implements Stage.
type StageFunc func(ctx context.Context, l *lease.Lease) (*lease.Lease, error)

// Process implements Stage.
func (f StageFunc) Process(ctx context.Context, l *lease.Lease) (*lease.Lease, error) {
	return f(ctx, l)
}

// ErrSkip can be returned by a Stage to indicate that the lease should be
// silently dropped. The pipeline treats this as a non-error early exit.
var ErrSkip = fmt.Errorf("pipeline: lease skipped")

// Pipeline runs a lease through an ordered sequence of Stages. Processing
// stops at the first error; if the error is ErrSkip the lease is dropped
// without surfacing an error to the caller.
type Pipeline struct {
	stages []Stage
}

// New creates a Pipeline with the provided stages executed in order.
func New(stages ...Stage) *Pipeline {
	s := make([]Stage, len(stages))
	copy(s, stages)
	return &Pipeline{stages: s}
}

// Append returns a new Pipeline with additional stages appended after the
// existing ones. The original Pipeline is not modified.
func (p *Pipeline) Append(stages ...Stage) *Pipeline {
	next := make([]Stage, len(p.stages)+len(stages))
	copy(next, p.stages)
	copy(next[len(p.stages):], stages)
	return &Pipeline{stages: next}
}

// Run executes all stages in order for the given lease. It returns nil if the
// lease was processed successfully or silently skipped. Any non-skip error is
// returned to the caller.
func (p *Pipeline) Run(ctx context.Context, l *lease.Lease) error {
	current := l
	for _, stage := range p.stages {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		out, err := stage.Process(ctx, current)
		if err != nil {
			if err == ErrSkip {
				return nil
			}
			return err
		}
		current = out
	}
	return nil
}

// RunAll executes the pipeline for each lease in the slice. Processing
// continues across leases even if one returns an error; all errors are
// collected and returned as a combined error. Skipped leases are not counted
// as errors.
func (p *Pipeline) RunAll(ctx context.Context, leases []*lease.Lease) error {
	var errs []error
	for _, l := range leases {
		if err := p.Run(ctx, l); err != nil {
			errs = append(errs, fmt.Errorf("lease %s: %w", l.ID, err))
		}
	}
	if len(errs) == 0 {
		return nil
	}
	return &MultiError{Errors: errs}
}

// MultiError holds one or more errors produced during a RunAll call.
type MultiError struct {
	Errors []error
}

// Error implements the error interface.
func (m *MultiError) Error() string {
	if len(m.Errors) == 1 {
		return m.Errors[0].Error()
	}
	return fmt.Sprintf("pipeline: %d errors occurred (first: %s)", len(m.Errors), m.Errors[0])
}

// Unwrap returns the first error for errors.Is / errors.As compatibility.
func (m *MultiError) Unwrap() error {
	if len(m.Errors) == 0 {
		return nil
	}
	return m.Errors[0]
}

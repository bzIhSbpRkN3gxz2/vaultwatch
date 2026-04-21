package tokenstore

import (
	"context"
	"errors"
)

type contextKey struct{}

// WithContext returns a new context carrying the given Store.
func WithContext(ctx context.Context, s *Store) context.Context {
	return context.WithValue(ctx, contextKey{}, s)
}

// FromContext retrieves a Store from ctx. Returns an error if none is present.
func FromContext(ctx context.Context) (*Store, error) {
	v := ctx.Value(contextKey{})
	if v == nil {
		return nil, errors.New("tokenstore: no store in context")
	}
	s, ok := v.(*Store)
	if !ok {
		return nil, errors.New("tokenstore: invalid store type in context")
	}
	return s, nil
}

// Lookup is a convenience helper that retrieves the store from ctx and calls
// Get. It propagates ErrNotFound and ErrExpired unchanged.
func Lookup(ctx context.Context, leaseID string) (*Entry, error) {
	s, err := FromContext(ctx)
	if err != nil {
		return nil, err
	}
	return s.Get(leaseID)
}

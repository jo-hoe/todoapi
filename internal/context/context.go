// Package context provides context utilities for the application
package context

import (
	"context"
	"time"
)

// WithTimeout creates a context with a timeout
func WithTimeout(parent context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if parent == nil {
		parent = context.Background()
	}
	return context.WithTimeout(parent, timeout)
}

// WithDefaultTimeout creates a context with a default 30-second timeout
func WithDefaultTimeout(parent context.Context) (context.Context, context.CancelFunc) {
	return WithTimeout(parent, 30*time.Second)
}

// Background returns a non-nil, empty Context
func Background() context.Context {
	return context.Background()
}

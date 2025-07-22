// Package common provides common utilities used across the application
package common

import (
	"io"
	"log/slog"
)

// CloseBody safely closes an io.Closer and logs any error
func CloseBody(c io.Closer) {
	if c == nil {
		return
	}

	if err := c.Close(); err != nil {
		slog.Error("failed to close body", "error", err)
	}
}

// CloseBodyWithCallback closes an io.Closer and calls the provided callback on error
func CloseBodyWithCallback(c io.Closer, onError func(error)) {
	if c == nil {
		return
	}

	if err := c.Close(); err != nil && onError != nil {
		onError(err)
	}
}

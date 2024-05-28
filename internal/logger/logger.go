// Package logger provides a function to create a new logger instance.
// It takes a logging level as a parameter and returns a pointer to a zap.Logger instance.
package logger

import (
	"go.uber.org/zap"
)

// New creates a new logger.
//
// Parameters:
//   - level: A string representing the logging level (e.g., "debug", "info", "warn", "error").
//
// Returns:
//   - *zap.Logger: A pointer to the created zap.Logger instance.
//   - error: An error if the logger could not be created, otherwise nil.
func New(level string) (*zap.Logger, error) {
	// Convert the textual logging level to zap.AtomicLevel
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return nil, err
	}
	// Create a new logger configuration
	cfg := zap.NewProductionConfig()
	// Set the logging level
	cfg.Level = lvl
	// Build the logger based on the configuration
	zl, err := cfg.Build()
	if err != nil {
		return nil, err
	}

	return zl, nil
}

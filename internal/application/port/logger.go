// Package port defines application-layer interfaces for infrastructure concerns.
// Implementations live in the infrastructure layer.
package port

// Logger provides leveled logging for application components.
// No Fatal method - termination decisions belong to main.go.
type Logger interface {
	// Debug logs a message at DEBUG level (verbose, development info).
	Debug(msg string, args ...any)

	// Info logs a message at INFO level (normal operational info).
	Info(msg string, args ...any)

	// Warn logs a message at WARN level (potential issues).
	Warn(msg string, args ...any)

	// Error logs a message at ERROR level (failures requiring attention).
	Error(msg string, args ...any)
}

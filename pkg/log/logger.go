// Package log provides a structured logging abstraction used across the application.
// The Logger interface is the only logging dependency that other packages should import.
// Concrete implementations live in this package and are created via constructors (e.g. NewZerologLogger).
package log

// FieldLogger represents a single structured key-value field attached to a log entry.
type FieldLogger struct {
	Key   string
	Value any
}

// Logger is the application-wide structured logging interface.
// All adapters, use-cases and transport layers depend on this interface, never on concrete implementations.
type Logger interface {
	Info(msg string, fields ...FieldLogger)
	Error(msg string, fields ...FieldLogger)
	Debug(msg string, fields ...FieldLogger)
	Warn(msg string, fields ...FieldLogger)
}

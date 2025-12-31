// Package logger provides a context-aware logging infrastructure
//
// # It supports leveled logging (DEBUG, INFO, WARN, ERROR, FATAL) and custom output formatting
//
// The logger is designed to be passed a context.Context, allowing for request-scoped or
// operation-scoped logging context to be maintained (though currently primarily used for
// cancellation propagation, it sets the stage for more structured logging)
package logger

package errors

import (
	"fmt"
	"runtime"

	"github.com/sirupsen/logrus"
)

// ScanError represents an error during malware scanning
type ScanError struct {
	Stage   string                 // Stage where error occurred: "validation", "plugin", "storage"
	File    string                 // File path or hash
	Code    string                 // Error code for categorization
	Message string                 // Human-readable message
	Err     error                  // Underlying error
	Context map[string]interface{} // Additional context
}

// Error implements the error interface
func (e *ScanError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s/%s] %s: %v", e.Stage, e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s/%s] %s", e.Stage, e.Code, e.Message)
}

// Unwrap allows error wrapping
func (e *ScanError) Unwrap() error {
	return e.Err
}

// ValidationError represents a validation failure
type ValidationError struct {
	Field   string
	Message string
	Value   interface{}
}

// Error implements the error interface
func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error on field %q: %s (value: %v)", e.Field, e.Message, e.Value)
}

// PluginError represents a plugin execution error
type PluginError struct {
	PluginName string
	ScanID     string
	Message    string
	Err        error
	ExitCode   int
}

// Error implements the error interface
func (e *PluginError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("plugin %q (scan %q): %s: %v", e.PluginName, e.ScanID, e.Message, e.Err)
	}
	return fmt.Sprintf("plugin %q (scan %q): %s (exit code: %d)", e.PluginName, e.ScanID, e.Message, e.ExitCode)
}

// Unwrap allows error wrapping
func (e *PluginError) Unwrap() error {
	return e.Err
}

// NewScanError creates a new ScanError
func NewScanError(stage, code, message string, err error) *ScanError {
	return &ScanError{
		Stage:   stage,
		Code:    code,
		Message: message,
		Err:     err,
		Context: make(map[string]interface{}),
	}
}

// NewValidationError creates a new ValidationError
func NewValidationError(field, message string, value interface{}) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: message,
		Value:   value,
	}
}

// NewPluginError creates a new PluginError
func NewPluginError(pluginName, scanID, message string, err error) *PluginError {
	return &PluginError{
		PluginName: pluginName,
		ScanID:     scanID,
		Message:    message,
		Err:        err,
	}
}

// CheckError logs an error if present (for compatibility with old code)
// Deprecated: Return errors explicitly instead of logging them
func CheckError(err error) bool {
	return CheckErrorWithMessage(err, "")
}

// CheckErrorNoStack logs an error without stack trace
// Deprecated: Return errors explicitly instead of logging them
func CheckErrorNoStack(err error) bool {
	return CheckErrorNoStackWithMessage(err, "")
}

// CheckErrorWithMessage logs an error with a message and stack trace
// Deprecated: Return errors explicitly instead of logging them
func CheckErrorWithMessage(err error, msg string, args ...interface{}) bool {
	if err != nil {
		var stack [4096]byte
		runtime.Stack(stack[:], false)

		if len(args) == 0 {
			logrus.WithError(err).Error(msg)
		} else {
			logrus.WithError(err).Errorf(msg, args...)
		}
		return false
	}
	return true
}

// CheckErrorNoStackWithMessage logs an error with a message but no stack trace
// Deprecated: Return errors explicitly instead of logging them
func CheckErrorNoStackWithMessage(err error, msg string, args ...interface{}) bool {
	if err != nil {
		if len(args) == 0 {
			logrus.WithError(err).Error(msg)
		} else {
			logrus.WithError(err).Errorf(msg, args...)
		}
		return false
	}
	return true
}

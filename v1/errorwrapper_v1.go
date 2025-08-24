package errorwrapper

import (
	"errors"
	"strings"
)

const (
	defaultErrJoiner byte   = 0x2E // Default character to join prefixes, which is '.'
	defaultMsgJoiner string = ": " // Default string to join prefixes and messages
)

// errorString is a simple struct that satisfies the error interface.
type errorString struct {
	s string
}

// Error implements the error interface for errorString.
func (e *errorString) Error() string {
	return e.s
}

// ErrorWrapper defines the interface for creating and wrapping errors.
type ErrorWrapper interface {
	// NewError wraps an existing error with a new message.
	NewError(msg string, err error) error
	// NewErrorString creates a new error from a string and wraps it with a message.
	NewErrorString(msg string, errStr string) error
}

// errWrapper is the concrete implementation of the ErrorWrapper interface.
type errWrapper struct {
	err       error
	msg       string
	prefix    string
	errJoiner byte
}

// Statically assert that *errWrapper implements the ErrorWrapper interface.
// This line will cause a compile-time error if the interface is not satisfied.
var _ ErrorWrapper = (*errWrapper)(nil)

// New creates and returns a new ErrorWrapper.
// It accepts an optional joiner byte for prefixes and an optional initial prefix string.
func New(errJoiner byte, prefix ...string) ErrorWrapper {
	ew := &errWrapper{
		errJoiner: errJoiner,
	}
	if ew.errJoiner == 0 {
		ew.errJoiner = defaultErrJoiner
	}
	if len(prefix) >= 1 {
		ew.prefix = prefix[0]
	}
	return ew
}

// NewError wraps an existing error with the wrapper's prefix and a new message.
// If the error being wrapped is also an errWrapper, it combines their prefixes.
func (ew errWrapper) NewError(msg string, err error) error {
	if errors.As(err, &errWrapper{}) {
		if j, exists := err.(*errWrapper); exists && j.prefix != "" {
			if ew.prefix == "" {
				return &errWrapper{
					prefix: j.prefix,
					msg:    msg,
					err:    j.err,
				}
			}
			var sb strings.Builder
			sb.WriteString(ew.prefix)
			sb.WriteByte(ew.errJoiner)
			sb.WriteString(j.prefix)
			return &errWrapper{
				prefix: sb.String(),
				msg:    msg,
				err:    j.err,
			}
		}
	}
	return &errWrapper{
		prefix: ew.prefix,
		err:    err,
		msg:    msg,
	}
}

// NewErrorString wraps a new error, created from a string, with the wrapper's prefix and a message.
func (ew errWrapper) NewErrorString(msg, errString string) error {
	return &errWrapper{
		prefix: ew.prefix,
		err:    &errorString{errString},
		msg:    msg,
	}
}

// Error implements the error interface for errWrapper, formatting the output string.
func (ew errWrapper) Error() string {
	var (
		sb          strings.Builder
		isMsgFilled bool = ew.msg != ""
	)
	if ew.prefix != "" {
		sb.WriteString(ew.prefix)
		sb.WriteString(defaultMsgJoiner)
	}
	if isMsgFilled {
		sb.WriteByte(0x5B)
		sb.WriteString(ew.msg)
		sb.WriteByte(0x5D)
	}
	if ew.err != nil {
		if isMsgFilled {
			sb.WriteByte(0x20)
		}
		sb.WriteString(ew.err.Error())
	}
	return sb.String()
}

// Unwrap returns the underlying wrapped error, allowing for error chain inspection.
func (ew *errWrapper) Unwrap() error {
	return ew.err
}

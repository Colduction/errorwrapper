package errorwrapper

import "strings"

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
	NewError(err error, msg ...string) error
	// NewErrorString creates a new error from a string and wraps it with a message.
	NewErrorString(errStr string, msg ...string) error
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

// unwrapRecursively traverses a chain of errWrapper errors.
// It returns the combined prefix of all wrappers and the root, non-wrapper error.
func unwrapRecursively(err error, joiner byte) (string, error) {
	if ew, ok := err.(*errWrapper); ok {
		recursivePrefix, underlyingErr := unwrapRecursively(ew.err, joiner)
		var sb strings.Builder
		sb.WriteString(ew.prefix)
		if ew.prefix != "" && recursivePrefix != "" {
			sb.WriteByte(joiner)
		}
		sb.WriteString(recursivePrefix)
		return sb.String(), underlyingErr
	}
	return "", err
}

// NewError wraps an existing error with the wrapper's prefix and a new message.
// If the error being wrapped is also an errWrapper, it combines their prefixes.
func (ew errWrapper) NewError(err error, msg ...string) error {
	if err == nil {
		return nil
	}
	var tmpMsg string
	if len(msg) >= 1 {
		tmpMsg = msg[0]
	}
	var (
		unwPrefix, undErr = unwrapRecursively(err, ew.errJoiner)
		sb                strings.Builder
	)
	sb.WriteString(ew.prefix)
	if ew.prefix != "" && unwPrefix != "" {
		sb.WriteByte(ew.errJoiner)
	}
	sb.WriteString(unwPrefix)
	return &errWrapper{
		prefix:    sb.String(),
		err:       undErr,
		msg:       tmpMsg,
		errJoiner: ew.errJoiner,
	}
}

// NewErrorString wraps a new error, created from a string, with the wrapper's prefix and a message.
func (ew errWrapper) NewErrorString(errStr string, msg ...string) error {
	if errStr == "" {
		return nil
	}
	var tmpMsg string
	if len(msg) >= 1 {
		tmpMsg = msg[0]
	}
	return &errWrapper{
		prefix: ew.prefix,
		err:    &errorString{errStr},
		msg:    tmpMsg,
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

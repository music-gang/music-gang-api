package apperr

import (
	"errors"
	"fmt"
)

// Application error codes.
//
// NOTE: These are meant to be generic and they map well to HTTP error codes.
// Different applications can have very different error code requirements so
// these should be expanded as needed (or introduce subcodes).
const (
	ECONFLICT       = "conflict"        // conflict with current state
	EINTERNAL       = "internal"        // internal error
	EINVALID        = "invalid"         // invalid input
	ENOTFOUND       = "not_found"       // resource not found
	ENOTIMPLEMENTED = "not_implemented" // feature not implemented
	EUNAUTHORIZED   = "unauthorized"    // access denied
	EUNKNOWN        = "unknown"         // unknown error
	EFORBIDDEN      = "forbidden"       // access forbidden
	EEXISTS         = "exists"          // resource already exists

	EMGVM                     = "mgvm"                // error code prefix for music gang virtual machine, it is assimilated to EINTERNAL
	EMGVM_LOWFUEL             = "low_fuel"            // subcode for EMGVM, low fuel
	EMGVM_CORE_POOL_NOT_FOUND = "core_pool_not_found" // subcode for EMGVM, core pool not found
	EMGVM_CORE_POOL_TIMEOUT   = "core_pool_timeout"   // subcode for EMGVM, core pool timeout

	EANCHORAGE = "anchorage" // error code prefix for anchorage contract executor, it is assimilated to EINTERNAL
)

// Error represents an application-specific error. Application errors can be
// unwrapped by the caller to extract out the code & message.
//
// Any non-application error (such as a disk error) should be reported as an
// EINTERNAL error and the human user should only see "Internal error" as the
// message. These low-level internal error details should only be logged and
// reported to the operator of the application (not the end user).
type Error struct {
	// Machine-readable error code.
	Code string

	// Human-readable error message.
	Message string
}

// Error implements the error interface.
func (e *Error) Error() string {
	return fmt.Sprintf("error: code=%s message=%s", e.Code, e.Message)
}

// ErrorCode unwraps an application error and returns its code.
// Non-application errors always return EINTERNAL.
func ErrorCode(err error) string {
	var e *Error
	if err == nil {
		return ""
	} else if errors.As(err, &e) {
		return e.Code
	}
	return EINTERNAL
}

// ErrorMessage unwraps an application error and returns its message.
// Non-application errors always return "Internal error".
func ErrorMessage(err error) string {
	var e *Error
	if err == nil {
		return ""
	} else if errors.As(err, &e) {
		return e.Message
	}
	return err.Error()
}

// ErrorLog returns the params for a log entry for an application error.
func ErrorLog(err error) (string, string, string) {
	var e *Error
	if err == nil {
		return "", "", ""
	} else if errors.As(err, &e) {
		return e.Message, "code", e.Code
	}
	return err.Error(), "code", EINTERNAL
}

// Errorf is a helper function to return an Error with a given code and formatted message.
func Errorf(code string, format string, args ...interface{}) *Error {
	return &Error{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
	}
}

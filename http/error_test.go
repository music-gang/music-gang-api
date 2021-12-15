package http_test

import (
	"testing"

	"github.com/music-gang/music-gang-api/app/apperr"
	"github.com/music-gang/music-gang-api/http"
)

func TestError_MessageFromErr(t *testing.T) {

	t.Run("OK", func(t *testing.T) {

		if err := http.MessageFromErr(apperr.Errorf(apperr.ENOTFOUND, "not found")); err != "not found" {
			t.Errorf("error: %v", err)
		}
	})

	t.Run("EmptyErr", func(t *testing.T) {

		if err := http.MessageFromErr(nil); err == "" {
			t.Error("error: expected error")
		} else if err != http.GenericErrorMessage {
			t.Errorf("error, expected: %s", http.GenericErrorMessage)
		}
	})

	t.Run("ErrInternal", func(t *testing.T) {

		if err := http.MessageFromErr(apperr.Errorf(apperr.EINTERNAL, "internal error")); err != http.DefaultInternalErrorMessage {
			t.Errorf("error, expected: %s", http.DefaultInternalErrorMessage)
		}
	})
}

func TestError_StatusCodeFromErr(t *testing.T) {

	t.Run("Conflict", func(t *testing.T) {
		if code := http.StatusCodeFromErr(apperr.Errorf(apperr.ECONFLICT, "conflict")); code != http.Codes[apperr.ECONFLICT] {
			t.Errorf("error, expected: %d", http.Codes[apperr.ECONFLICT])
		}
	})

	t.Run("Forbidden", func(t *testing.T) {
		if code := http.StatusCodeFromErr(apperr.Errorf(apperr.EFORBIDDEN, "forbidden")); code != http.Codes[apperr.EFORBIDDEN] {
			t.Errorf("error, expected: %d", http.Codes[apperr.EFORBIDDEN])
		}
	})

	t.Run("Internal", func(t *testing.T) {
		if code := http.StatusCodeFromErr(apperr.Errorf(apperr.EINTERNAL, "internal error")); code != http.Codes[apperr.EINTERNAL] {
			t.Errorf("error, expected: %d", http.Codes[apperr.EINTERNAL])
		}
	})

	t.Run("Invalid", func(t *testing.T) {
		if code := http.StatusCodeFromErr(apperr.Errorf(apperr.EINVALID, "invalid")); code != http.Codes[apperr.EINVALID] {
			t.Errorf("error, expected: %d", http.Codes[apperr.EINVALID])
		}
	})

	t.Run("NotFound", func(t *testing.T) {
		if code := http.StatusCodeFromErr(apperr.Errorf(apperr.ENOTFOUND, "not found")); code != http.Codes[apperr.ENOTFOUND] {
			t.Errorf("error, expected: %d", http.Codes[apperr.ENOTFOUND])
		}
	})

	t.Run("NotImplemented", func(t *testing.T) {
		if code := http.StatusCodeFromErr(apperr.Errorf(apperr.ENOTIMPLEMENTED, "not implemented")); code != http.Codes[apperr.ENOTIMPLEMENTED] {
			t.Errorf("error, expected: %d", http.Codes[apperr.ENOTIMPLEMENTED])
		}
	})

	t.Run("Unauthorized", func(t *testing.T) {
		if code := http.StatusCodeFromErr(apperr.Errorf(apperr.EUNAUTHORIZED, "unauthorized")); code != http.Codes[apperr.EUNAUTHORIZED] {
			t.Errorf("error, expected: %d", http.Codes[apperr.EUNAUTHORIZED])
		}
	})

	t.Run("GenericError", func(t *testing.T) {
		if code := http.StatusCodeFromErr(apperr.Errorf("generic", "generic error")); code != http.Codes[apperr.EINTERNAL] {
			t.Errorf("error, expected: %d", http.Codes[apperr.EINTERNAL])
		}
	})
}

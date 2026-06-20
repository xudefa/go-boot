package exception

import (
	"errors"
	"testing"
)

func TestBuiltinErrors(t *testing.T) {
	if ErrNotFound == nil {
		t.Error("ErrNotFound should not be nil")
	}
	if ErrBadRequest == nil {
		t.Error("ErrBadRequest should not be nil")
	}
	if ErrUnauthorized == nil {
		t.Error("ErrUnauthorized should not be nil")
	}
	if ErrForbidden == nil {
		t.Error("ErrForbidden should not be nil")
	}
	if ErrConflict == nil {
		t.Error("ErrConflict should not be nil")
	}
	if ErrInternalServer == nil {
		t.Error("ErrInternalServer should not be nil")
	}

	if !errors.Is(ErrNotFound, ErrNotFound) {
		t.Error("ErrNotFound should be equal to itself")
	}
}

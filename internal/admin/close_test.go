package admin

import (
	"errors"
	"testing"
)

type fakeCloser struct {
	err error
}

func (f fakeCloser) Close() error {
	return f.err
}

func TestCloseWithErrorReturnsCloseErrorAfterSuccess(t *testing.T) {
	closeErr := errors.New("close failed")
	var resultErr error

	closeWithError(fakeCloser{err: closeErr}, &resultErr)

	if !errors.Is(resultErr, closeErr) {
		t.Fatalf("expected close error, got %v", resultErr)
	}
}

func TestCloseWithErrorPreservesOperationAndCloseErrors(t *testing.T) {
	operationErr := errors.New("operation failed")
	closeErr := errors.New("close failed")
	resultErr := operationErr

	closeWithError(fakeCloser{err: closeErr}, &resultErr)

	if !errors.Is(resultErr, operationErr) {
		t.Fatalf("expected operation error to be preserved, got %v", resultErr)
	}
	if !errors.Is(resultErr, closeErr) {
		t.Fatalf("expected close error to be preserved, got %v", resultErr)
	}
}

func TestCloseWithErrorIgnoresNilArguments(t *testing.T) {
	var resultErr error
	closeWithError(nil, &resultErr)
	closeWithError(fakeCloser{}, nil)

	if resultErr != nil {
		t.Fatalf("expected nil error, got %v", resultErr)
	}
}

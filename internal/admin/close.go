package admin

import "errors"

type errorCloser interface {
	Close() error
}

func closeWithError(closer errorCloser, resultErr *error) {
	if closer == nil || resultErr == nil {
		return
	}
	*resultErr = errors.Join(*resultErr, closer.Close())
}

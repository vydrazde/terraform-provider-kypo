package KYPOClient

import "errors"

var ErrTimeout = errors.New("has not finished within timeout")

var ErrNotFound = errors.New("not found")

type valueOrError[T any] struct {
	err   error
	value T
}

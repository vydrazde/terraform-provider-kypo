package KYPOClient

import (
	"fmt"
	"time"
)

type ErrTimeout struct {
	Action     string
	Identifier string
	Timeout    time.Duration
}

func (e *ErrTimeout) Error() string {
	return fmt.Sprintf("%s: %s has not finished within %s", e.Action, e.Identifier, e.Timeout)
}

type ErrNotFound struct {
	ResourceName string
	Identifier   string
}

func (e *ErrNotFound) Error() string {
	return fmt.Sprintf("resource %s: %s was not found", e.ResourceName, e.Identifier)
}

type valueOrError[T any] struct {
	err   error
	value T
}

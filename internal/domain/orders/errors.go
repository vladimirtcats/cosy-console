package orders

import "errors"

var (
	ErrNotFound          = errors.New("order not found")
	ErrInvalidTransition = errors.New("order status transition is not allowed")
)

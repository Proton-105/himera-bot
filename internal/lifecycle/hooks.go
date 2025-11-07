package lifecycle

import "context"

// Hook describes a named shutdown hook.
type Hook struct {
	Name string
	Fn   func(ctx context.Context) error
}

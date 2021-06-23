package references

import (
	"context"
	"time"
)

func definingContext() {
	// There's two ways to define a brand-new context
	// Background() gives us an empty context with absolutely nothing.
	// This is typically used in the main function for intialization purposes.
	ctx := context.Background()
	// This also provides an empty context, but it's used as a placeholder for
	// when it's uncleare which context to use or when a parent context is
	// not yet available.
	context.TODO()

	// There are other ways to define context with cancellations and timeouts.
	// WithDeadline accepts a specific date the context will be cancelled
	context.WithDeadline(ctx, time.Now().AddDate(0, 0, 1))
	// WithTimeout accepts a duration
	context.WithTimeout(ctx, time.Second*10)
	// WithCancel just gives you a cancel function you can call to explicitly cancel the context
	// Keep in mind that the other two contexts give you this function as well
	context.WithCancel(ctx)
	// You can also initialize a key-value pair inside a context if you want to use request-scoped
	// values, but this is typically not done.
	context.WithValue(ctx, "key", "value")
}

// Create a function that performs a slow operation
// The out channel should receive a value from a slow operation
// The function should time out after 10 seconds if the slow operation doesn't complete
func stream(ctx context.Context, out chan<- int) error {

	dctx, cancel := context.WithTimeout(ctx, time.Second*10)
	// We will call cancel() if the slow operation completes before timeout
	defer cancel()

	resp, err := slowOperation(dctx)
	if err != nil {
		return err
	}

	for {
		select {
		case out <- resp:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// This operation isn't actually slow but let's just pretend that it is
func slowOperation(ctx context.Context) (int, error) {

	return 0, nil
}

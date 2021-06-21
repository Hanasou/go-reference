package main

import (
	"context"
	"errors"
	"sync"
	"time"
)

// The circuit function type represents the function that interacts with the service
// Think about a function that makes a database call or something like that
// We're defining Circuit as a type here to denote that we want to pass in any function
// that meets the criteria for this type into our breaker function
type Circuit func(context.Context) (string, error)

// Effector is actually the exact same as Circuit, but we're using it in a different context
type Effector func(context.Context) (string, error)

// Breaker is a function that that will retry our circuit a certain number of times.
// If it keeps failing, the, we will return a new error.
// We're passing in a function into another function to basically add some additional functionality to it
func Breaker(circuit Circuit, failureThreshold uint) Circuit {
	var consecutiveFailures int = 0
	var lastAttempt = time.Now()
	var m sync.RWMutex

	return func(ctx context.Context) (string, error) {
		// Establish read lock
		// We need read locks because we're reading external state
		m.RLock()

		d := consecutiveFailures - int(failureThreshold)
		// If we've passed the failure threshold
		if d >= 0 {
			shouldRetryAt := lastAttempt.Add(time.Second * 2 << d) // backoff logic
			if !time.Now().After(shouldRetryAt) {
				m.RUnlock()
				return "", errors.New("service unreachable")
			}
		}

		m.RUnlock()

		response, err := circuit(ctx) // Call the parent function

		// Lock resources
		m.Lock()
		defer m.Unlock()

		lastAttempt = time.Now()

		// If the circuit function returned an error
		if err != nil {
			consecutiveFailures++
			return response, err
		}

		// If the function succeeded, reset consecutiveFailures
		consecutiveFailures = 0

		return response, nil
	}
}

// Like with Breaker, the Debounce function will also take in a circuit and add some
// additional functionality to it.
// This time, we want to ignore all the middle function calls in a cluster of function calls.
// We will implement this by adding a time duration for each function call
// Any function call that follows after this interval will be ignored
func Debounce(circuit Circuit, duration time.Duration) Circuit {
	var threshold time.Time
	var result string
	var err error
	var m sync.Mutex

	return func(ctx context.Context) (string, error) {
		m.Lock() // We're using external state here so we have to lock

		// Determine threshold here
		defer func() {
			threshold = time.Now().Add(duration)
			m.Unlock()
		}()

		// If the function is called before the threshold
		// Get the previous result/response
		if time.Now().Before(threshold) {
			return result, err
		}
		result, err = circuit(ctx)

		return result, nil
	}
}

// We want to add some additional functionality to our Effector function type
// Allow our function to automatically retry a certain number of times
func Retry(effector Effector, retries int, delay time.Duration) Effector {
	return func(ctx context.Context) (string, error) {
		// This for loop lacks an upper bound. We'll take care of this later
		for r := 0; ; r++ {
			resp, err := effector(ctx)
			// This represents the success case
			if err == nil || r >= retries {
				return resp, nil
			}
			select {
			case <-time.After(delay): // We block on this case until after the delay
			case <-ctx.Done():
				return "", ctx.Err()
			}
		}
	}
}

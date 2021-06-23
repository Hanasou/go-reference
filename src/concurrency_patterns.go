package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Fan-in pattern should multiplex multiple input channels to one output channel
// Accept a source of multiple receiver channels
func Funnel(sources ...<-chan int) <-chan int {
	// shared output channel
	dest := make(chan int)
	// We use wait groups to sync up goroutines
	var wg sync.WaitGroup

	// Number of times Done() needs to get called on the Wait Group
	// is the number of sources we input into this function
	wg.Add(len(sources))

	// Start a goroutine for each source
	for _, ch := range sources {
		go func(c <-chan int) {
			// Call this when the goroutine finishes
			defer wg.Done()
			// n is an integer our source channels receive
			// Put that into our output channel when we receive it
			// We're closing our input channel in our run function
			// Basically whenever we range over a channel, we need to close it manually
			for n := range c {
				dest <- n
			}
		}(ch)
	}

	// We're starting another goroutine here for our wait group to block
	// until all source channels are done and to close dest.
	// We need to close dest because we're ranging over it in our run function
	go func() {
		wg.Wait()
		close(dest)
	}()

	return dest
}

// How we start our fan-in simulation
func runFanIn() {
	// Make a slice of input int channels
	sources := make([]<-chan int, 0)

	for i := 0; i < 3; i++ {
		// Create a channel and add to sources
		ch := make(chan int)
		sources = append(sources, ch)

		// Run a goroutine to simulate some work
		go func() {
			// Close when we're done sending stuff to goroutine
			defer close(ch)
			// Send ints to our input channel
			for i := 0; i < 5; i++ {
				ch <- i
				time.Sleep(time.Second)
			}
		}()
	}

	dest := Funnel(sources...)
	for d := range dest {
		fmt.Println(d)
	}
}

// Fan-out pattern will take one input channel and have multiple
// output channels read from it.
func Split(source <-chan int, n int) []<-chan int {
	// Create destinations slice
	dests := make([]<-chan int, 0)

	// Create destination channels
	for i := 0; i < n; i++ {
		ch := make(chan int)      // make int channel
		dests = append(dests, ch) // append channel to dests
		// Make a goroutine for each output channel
		// Read data from input channel
		go func() {
			defer close(ch)
			// Sender will have to close source
			for val := range source {
				ch <- val
			}
		}()
	}

	return dests
}

func runFanOut() {
	source := make(chan int)  // make source channel
	dests := Split(source, 5) // make five destination channels

	// produce data to source
	go func() {
		for i := 0; i < 10; i++ {
			source <- i
		}
		close(source)
	}()

	var wg sync.WaitGroup // use a wait group to sync our output goroutines
	wg.Add(len(dests))

	// read data from dests
	for i, ch := range dests {
		go func(i int, d <-chan int) {
			// We closed our output channels inside the Split function
			for val := range d {
				fmt.Printf("#%d got %d\n", i, val)
			}
		}(i, ch)
	}

	wg.Wait()
}

// Create some types that are involved with the Future pattern
// Generally you can achieve the same sort of thing with just channels
// and goroutines, but instead we can encapsulate these things
// with a high level api.
type Future interface {
	Result() (string, error)
}

type InnerFuture struct {
	once sync.Once
	wg   sync.WaitGroup

	res   string
	err   error
	resCh <-chan string
	errCh <-chan error
}

// We want to get the result of res and err from our Future
// We grab the results from channels.
// When Result() is first called, we get the value from the channels
// Then later we can access the res and err fields directly
func (f *InnerFuture) Result() (string, error) {
	// Ensure that this gets called only once
	f.once.Do(func() {
		// Wait for this thread to finish
		f.wg.Add(1)
		defer f.wg.Done()
		f.res = <-f.resCh
		f.err = <-f.errCh
	})

	f.wg.Wait()

	return f.res, f.err
}

// SlowFunction should contain the core functionality of what we're trying to do
// such as querying some data.
// It's responsible for creating the channels, running the core functionality in a goroutine,
// and creating and returning the Future implementation
func SlowFunction(ctx context.Context) Future {
	resCh := make(chan string)
	errCh := make(chan error)

	// Run the core functionality in a goroutine
	go func() {
		// The core functionality should send its response to the res channel
		// and error to the error channel
		select {
		case <-time.After(time.Second * 2):
			resCh <- "I slept for 2 seconds"
			errCh <- nil
		case <-ctx.Done():
			resCh <- ""
			errCh <- ctx.Err()
		}
	}()

	return &InnerFuture{resCh: resCh, errCh: errCh}
}

func runFuture() {
	ctx := context.Background()
	// Put responses into channels
	future := SlowFunction(ctx)
	// Put the results from channels into future
	res, err := future.Result()
	if err != nil {
		fmt.Println("error: ", err)
		return
	}

	fmt.Println(res)
}

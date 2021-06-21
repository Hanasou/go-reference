package main

import "fmt"

func channelBlocking() {
	ch := make(chan string) // Make a channel that receives strings

	// The go keyword before this function indicates that this will not block the main thread
	go func() {
		// When we receive on a channel, we block until we send a message to that channel
		// <-ch means we're receiving a value and assigning it to message
		message := <-ch
		fmt.Println(message)
		// Sending "pong" to channel
		// This will block until the channel receives the message
		ch <- "pong"
	}()

	// Since the above code is running on a separate goroutine, the below code will immediately run
	// After we make the channel

	ch <- "ping" // That being said, we shall immediately block it
	// See the code in the goroutine that's blocking on a receive
	// "ping" will get sent to the channel, and the message that's received on a separate goroutine
	// will be "ping".
	// The goroutine will then send the message "pong" to the channel
	// and we will receive it here.
	fmt.Println(<-ch) // This will print pong
}

func channelBuffering() {
	// If we initialize a channel with a buffer
	// The channel will only block on sends if the buffer is full
	// The channel will only block on receives if the buffer is empty
	// Therefore, if we don't initialize it with a buffer, you can think of the buffer size being 0
	ch := make(chan string, 2) // Buffer size is the second argument

	// These two sends will not block
	ch <- "foo"
	ch <- "bar"

	// These two receives will not block
	fmt.Println(<-ch)
	fmt.Println(<-ch)
	// Third receive will block (buffer is empty)
	fmt.Println(<-ch)
}

func channelLooping() {
	// We can also use the range keyword on channels to iterate over them
	// Ranging over channels will block until we either close the channel
	// or until all values are available to be read
	ch := make(chan string, 3)
	ch <- "foo"
	ch <- "bar"
	ch <- "baz"

	close(ch)

	// Range will continue up to "closed" flag
	for s := range ch {
		fmt.Println(s)
	}
}

package main

import (
	// System imports
	"os"
	"fmt"
	"sync"
	"os/signal"

	// go-efsw imports
	efsw "github.com/havoc-io/go-efsw"
)

func main() {
	// Check arguments
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: example <path> [<path>...] ")
        os.Exit(1)
	}

	// Create a wait group to monitor watcher threads
	var waitGroup sync.WaitGroup

	// Create a list of watches
	watches := make([]efsw.Watch, len(os.Args) - 1)

	// Loop over every path that we want to monitor and spawn a goroutine with
	// a watcher
	for i, path := range os.Args[1:] {
		// Add another entry to wait for
		waitGroup.Add(1)

		// Create the watch with a 10 event buffer
		watch := efsw.NewWatch(path, true, 10)

		// Record it
		watches[i] = watch

		// Go grab its events and print them out
		// NOTE: Have to pass path as an argument since it is an iteration
		// variable that won't be captured properly by closures (its last value
		// will be taken)
		go func (watchPath string) {
			// Grab events while they're coming in and the channel isn't closed
			for e := range watch.Events {
				fmt.Printf(
					"%s for %s (formerly %s) in %s\n",
					efsw.EventTypeToName[e.Type],
					e.Filename,
					e.OldFilename,
					e.Directory,
				)
			}

			// Say we're done
			fmt.Printf("Watcher for %s is finished\n", watchPath)

			// This watching thread is done
			waitGroup.Done()
		}(path)
	}

	// Watch for a control-C event, shutting down all watches and waiting for
	// completion before exiting
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt)
	for _ = range signalChannel {
		fmt.Println("Received interrupt, shutting down watchers...")
		for _, w := range watches {
			efsw.DeleteWatch(w)
		}
		waitGroup.Wait()
		fmt.Println("До свидания!")
		break
	}
}

package efsw

/*
#cgo pkg-config: efsw
#include <stdlib.h>
#include "go_efsw.h"
*/
import "C"

import (
	// System imports
	"sync"
	"unsafe"
)

// Event types (set to match efsw's API)
const (
	EventAdd = C.EFSW_ADD
	EventDelete = C.EFSW_DELETE
	EventModified = C.EFSW_MODIFIED
	EventMoved = C.EFSW_MOVED
)

// Map of event types to name
var EventTypeToName = map[int]string {
	EventAdd: "add",
	EventDelete: "delete",
	EventModified: "modified",
	EventMoved: "moved",
}

// Type representing filesystem events
type Event struct {
	Directory string
	Filename string
	Type int
	OldFilename string
}

// Type representing a watch
type Watch struct {
	watchId C.efsw_watchid
	Events chan Event
}

// Global watcher
var watcher C.efsw_watcher

// Global map from watch key to event channel (necessary since we can't
// associate a persistent (non-GC'd) pointer with the C callback)
var dispatchMap = make(map[C.efsw_watchid]chan Event)

// Global lock designed to coordinate access to the efsw API (which is not
// thread-safe and needs a write lock for use) and the watch-to-channel map
// (which requires a read lock for dispatch and a write lock for mutation).  We
// could use an RWMutex here, which would improve performance if we had many
// watcher threads dispatching events, but we only have one watcher thread
// locking it for reads, so there won't be any contention between readers
// anyway.
var lock sync.Mutex

// Watcher callback
//export watcherCallback
func watcherCallback(
	watchId C.efsw_watchid,
	directory,
	filename *C.char,
	action C.int,
	oldFilename *C.char) {
	// Lock the watch-to-channel map for reading
	lock.Lock()
	defer lock.Unlock()

	// Get the event channel, ignoring the callback if we can't find it, because
	// the channel may have been removed from the dispatch map by DeleteWatch
	// before the watcher thread managed to post the event
	channel := dispatchMap[watchId]
	if channel == nil {
		return
	}

	// Create the event
	event := Event{
		C.GoString(directory),
		C.GoString(filename),
		int(action),
		C.GoString(oldFilename)}

	// Dispatch the event if immediately possible, otherwise discard it
	select {
	case channel <- event:
	default:
	}
}

// Watch factory
func NewWatch(path string, recursive bool, buffer int) Watch {
	// Lock the API and watch-to-channel map for writing
	lock.Lock()
	defer lock.Unlock()

	// Ensure that the global watcher has been initialized
	if watcher == nil {
		// Allocate the watcher
		watcher = C.efsw_create(0)

		// Tell it to launch its watching thread
		C.efsw_watch(watcher)
	}

	// Convert the path
	pathCString := C.CString(path)
	defer C.free(unsafe.Pointer(pathCString))

	// Convert the recursive option
	recursiveCInt := C.int(1)
	if !recursive {
		recursiveCInt = 0
	}

	// Add the watch (we don't need to worry about it firing a callback before
	// the channel is in place because we're holding the lock at the moment)
	watchId := C.go_efsw_add_watch(
		watcher,
		pathCString,
		recursiveCInt)

	// Create the event channel
	channel := make(chan Event, buffer)

	// Record the channel in the dispatch map
	dispatchMap[watchId] = channel

	// All done
	return Watch{watchId, channel}
}

// Removes a watch
func DeleteWatch(watch Watch) {
	// Lock the API and watch-to-channel map for writing
	lock.Lock()
	defer lock.Unlock()

	// Remove the watch
	C.efsw_removewatch_byid(watcher, watch.watchId)

	// Remove the entry from the dispatch map
	delete(dispatchMap, watch.watchId)

	// Close the channel (will be fine since it's no longer in the dispatch map
	// and the watcher thread can't write to it)
	close(watch.Events)
}

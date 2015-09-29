# go-efsw

Go wrapper for efsw.

efsw is a cross-platform file system watcher and notifier with a broad range of
platform support available under the MIT license.  For more information, please
see the [efsw Bitbucket page](https://bitbucket.org/SpartanJ/efsw)<sup>1</sup>.

<sup>
1: At the moment, the primary efsw distribution doesn't generate a pkg-config
configuration file, which is necessary for cgo integration, so you'll have to
install from [this fork](https://bitbucket.org/havoc-io/efsw) using the CMake
build system.  I'll try to extend the pkg-config support to the Premake build
system and get it merged upstream at some point in the future.
</sup>


## Status

The module is currently supported (though not tested) on all platforms that efsw
supports.  If you run into trouble, please file an issue.  There are no tests at
the momemnt, mostly because the nature of file system notifications across
platforms is highly variable.  There is, however, an example program available
in the `example/` subdirectory.


## Dependencies

efsw does not have any third-party dependencies, and go-efsw automatically picks
up any system dependencies from efsw, so things *should* work out of the box.
Please file an issue if not.


## Usage

go-efsw provides a simplified (but still fully functional) subset of the efsw
interface, and is designed to follow Go's idioms.

efsw generally supports multiple watchers, each with multiple watches.  Each
watcher runs its own OS-level thread.  However, because the cgo interface does
not provide a nice interface for managing GC-driven deletion of C++ objects,
go-efsw simply uses a single global watcher/thread.  This is mostly just an
implementation detail - the API still allows you to create as many file/folder
watches as you like, supports recursive and non-recursive watches, is completely
thread-safe, and more than performant enough to handle whatever the OS throws at
it.  This design has the advantage of simplicity and low resource usage.

There is no need to initialize the global watcher.  It will be automatically
initialized the first time any watch is created.

To create a new watch object, simply do:

	path := "/Some/folder/to/watch"
	recursive := true
	buffer := 10
	watch, err := efsw.NewWatch(path, recursive, buffer)
	if err != nil {
		// Handle error
	}

Note that path needs to be a valid UTF-8 encoded string.  Additionally, it may
not contain null bytes.  The `efsw.NewWatch` function will validate this.  All
`efsw.Event` members will be UTF-8 encoded strings.

The `watch` object will contain a member named `Events`, which is of type
`chan efsw.Event`.  The channel will be given a buffer size as specified in the
third argument of `efsw.NewWatch`.  **Events that cannot be written to the
channel without blocking will be discarded.**  This means that you need to
digest events on the channel regularly, otherwise new ones will be discarded
until the channel has space to accomodate them.  Given that Go's scheduling
algorithm might schedule threads in such a way that a channel listener might not
be polling the channel exactly when the event is delivered, it's advised that
you use a non-0 buffer size (though you don't need a 1000-event buffer).

You can monitor events by doing something like:

	for e := range watch.Events {
		fmt.Printf(
			"%s for %s (formerly %s) in %s\n",
			efsw.EventTypeToName[e.Type],
			e.Filename,
			e.OldFilename,
			e.Directory,
		)
	}

You can stop a watch by doing:

	efsw.DeleteWatch(watch)

The `efsw.DeleteWatch` function will close the event channel, so it provides an
ideal mechanism to stop any Go routines that are polling for events with `range`
or similar.

If you don't call the `efsw.DeleteWatch` function, the watch will continue,
dropping events if they aren't pulled off the event channel.  This won't hurt
anything, but it will consume resources, so it's best to stop the watch when
it's not longer needed.


## Caveats

The standard caveats for any file system monitoring library apply:

- The events generated are highly dependent on the platform (e.g. some platforms
  and monitoring mechanisms don't support move, only add(copy)/delete)
- The events generated are highly dependent on the program causing the changes
  (e.g. some programs will save files to a temporary location and then
  atomically rename them)
- The features supported are highly dependent on the platform (e.g. some
  platforms only support directory monitoring, but won't give you an error if
  you try to monitor a file)
- Notification rates and times may vary, even on the same platform
- Some notifications may be dropped by the OS

Fortunately, efsw does a pretty good job of normalizing things, and it supports
recursive directory monitoring on all platforms.

**Your best bet, however, is to use go-efsw as a mechanism to indicate that you
should do some more precise inspection of the state of a given path**.  Don't
rely on file system notifications for fine grain tracking.  They are more like
the hoard of people running in a given direction with Cthulhu in pursuit - i.e.
you should follow them and be aware of the situation, but don't rely on them for
your survival.


## TODO:

- It might be nicer if watch event buffers dropped older events first, but such
  a "ring channel" is a bit complicated to implement in Go.  In any case, events
  should be serviced regularly, but perhaps it would be nice to have an option
  to control this behavior.

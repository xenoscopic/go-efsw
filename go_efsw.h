#ifndef GO_EFSW_H
#define GO_EFSW_H

// efsw includes
#include <efsw/efsw.h>

// Registers a new watch (necessary since we can't pass a function callback from
// inside Go)
efsw_watchid go_efsw_add_watch(
	efsw_watcher watcher,
	const char * path,
	int recursive
);

#endif /* GO_EFSW_H */

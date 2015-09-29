// cgo includes
#include "_cgo_export.h"

void go_efsw_watcher_callback(
	efsw_watcher watcher,
	efsw_watchid watchid,
	const char* dir,
	const char* filename,
	enum efsw_action action,
	const char* old_filename,
	void* param
) {
	// Call the Go callback
	watcherCallback(
		watchid,
		(char *)dir,
		(char *)filename,
		action,
		(char *)old_filename
	);
}

efsw_watchid go_efsw_add_watch(
	efsw_watcher watcher,
	const char * path,
	int recursive
) {
	return efsw_addwatch(
		watcher,
		path,
		go_efsw_watcher_callback,
		recursive,
		NULL
	);
}

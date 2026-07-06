package cacheindex

// WorkerLimit returns workers if positive, otherwise def. Backend
// implementations and the index updater use it to size their concurrency
// instead of hardcoding a limit.
func WorkerLimit(workers, def int) int {
	if workers > 0 {
		return workers
	}
	return def
}

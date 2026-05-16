package workflow

import "sync"

// RunBounded runs fn for each item with at most maxWorkers concurrent goroutines.
func RunBounded[T any](maxWorkers int, items []T, fn func(T)) {
	if len(items) == 0 {
		return
	}
	if maxWorkers < 1 {
		maxWorkers = 1
	}
	sem := make(chan struct{}, maxWorkers)
	var wg sync.WaitGroup
	for _, item := range items {
		wg.Add(1)
		go func(it T) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			fn(it)
		}(item)
	}
	wg.Wait()
}

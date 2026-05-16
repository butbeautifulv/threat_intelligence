package workflow

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestRunBounded_capsConcurrency(t *testing.T) {
	var peak, current int32
	RunBounded(2, []int{0, 1, 2, 3, 4}, func(_ int) {
		c := atomic.AddInt32(&current, 1)
		for {
			p := atomic.LoadInt32(&peak)
			if c <= p || atomic.CompareAndSwapInt32(&peak, p, c) {
				break
			}
		}
		time.Sleep(20 * time.Millisecond)
		atomic.AddInt32(&current, -1)
	})
	if peak > 2 {
		t.Fatalf("peak concurrency %d, want <= 2", peak)
	}
}

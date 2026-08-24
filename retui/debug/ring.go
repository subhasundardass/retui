package debug

import (
	"sync"
	"time"
)

// Entry represents a single captured log line, retained for inspection
// via Entries or rendering in an on-screen debug overlay.
type Entry struct {
	Time    time.Time
	Level   Level
	Message string
}

// ringBuffer is a fixed-capacity circular buffer of Entry values, safe
// for concurrent use.
type ringBuffer struct {
	mu   sync.Mutex
	buf  []Entry
	head int // next write index
	size int // number of valid entries currently stored
}

func newRingBuffer(capacity int) *ringBuffer {
	if capacity < 1 {
		capacity = 1
	}
	return &ringBuffer{buf: make([]Entry, capacity)}
}

func (r *ringBuffer) push(e Entry) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.buf[r.head] = e
	r.head = (r.head + 1) % len(r.buf)
	if r.size < len(r.buf) {
		r.size++
	}
}

// snapshot returns retained entries oldest-first.
func (r *ringBuffer) snapshot() []Entry {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.snapshotLocked()
}

// snapshotLocked assumes r.mu is already held.
func (r *ringBuffer) snapshotLocked() []Entry {
	out := make([]Entry, r.size)
	if r.size == 0 {
		return out
	}
	start := (r.head - r.size + len(r.buf)) % len(r.buf)
	for i := 0; i < r.size; i++ {
		out[i] = r.buf[(start+i)%len(r.buf)]
	}
	return out
}

// resize changes capacity, preserving as many of the most recent entries
// as fit in the new capacity.
func (r *ringBuffer) resize(capacity int) {
	if capacity < 1 {
		capacity = 1
	}
	r.mu.Lock()
	defer r.mu.Unlock()

	old := r.snapshotLocked()
	if len(old) > capacity {
		old = old[len(old)-capacity:]
	}
	r.buf = make([]Entry, capacity)
	r.head = 0
	r.size = 0
	for _, e := range old {
		r.buf[r.head] = e
		r.head = (r.head + 1) % len(r.buf)
		r.size++
	}
}

func (r *ringBuffer) clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.head = 0
	r.size = 0
}

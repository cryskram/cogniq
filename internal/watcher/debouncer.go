package watcher

import (
	"sync"
	"time"
)

type Debouncer struct {
	interval time.Duration
	callback func(paths []string)

	mu    sync.Mutex
	timer *time.Timer
	set   map[string]struct{}
}

func NewDebouncer(interval time.Duration, callback func(paths []string)) *Debouncer {
	return &Debouncer{
		interval: interval,
		callback: callback,
		set:      make(map[string]struct{}),
	}
}

func (d *Debouncer) Add(path string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.set[path] = struct{}{}

	if d.timer != nil {
		d.timer.Stop()
	}
	d.timer = time.AfterFunc(d.interval, d.fire)
}

func (d *Debouncer) fire() {
	d.mu.Lock()
	paths := make([]string, 0, len(d.set))
	for p := range d.set {
		paths = append(paths, p)
	}
	d.set = make(map[string]struct{})
	d.timer = nil
	d.mu.Unlock()

	if len(paths) > 0 {
		d.callback(paths)
	}
}

func (d *Debouncer) Stop() {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.timer != nil {
		d.timer.Stop()
		d.timer = nil
	}
	d.set = make(map[string]struct{})
}

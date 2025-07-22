package finalizer

import (
	"sync"
)

var (
	cleanupFuncs []func() error
	mut          sync.Mutex
)

func RegisterCleanupFuncs(f ...func() error) {
	mut.Lock()
	defer mut.Unlock()

	if f != nil {
		cleanupFuncs = append(cleanupFuncs, f...)
	}
}

func RunCleanupFuncs() []error {
	mut.Lock()
	defer mut.Unlock()

	e := make([]error, 0, len(cleanupFuncs))
	for i := len(cleanupFuncs) - 1; i >= 0; i-- {
		f := cleanupFuncs[i]
		if f != nil {
			if err := f(); err != nil {
				e = append(e, err)
			}
		}
	}
	return e
}

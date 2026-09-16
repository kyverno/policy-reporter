package report

import "sync"

// InitialReports tracks the first LIST delivered to each report handler and its
// successful persistence. Each reporting API owns a tracker, with namespace/name
// keys (bare names for cluster reports), retained across informer restarts.
// A nil tracker disables the initialization gate.
type InitialReports struct {
	mu      sync.Mutex
	sources map[string]bool
	pending map[string]struct{}
}

func NewInitialReports(sources ...string) *InitialReports {
	initial := &InitialReports{sources: make(map[string]bool), pending: make(map[string]struct{})}
	for _, source := range sources {
		initial.sources[source] = false
	}
	return initial
}

func (i *InitialReports) Track(source, key string) {
	if i == nil {
		return
	}
	i.mu.Lock()
	defer i.mu.Unlock()
	if sealed, exists := i.sources[source]; exists && !sealed {
		i.pending[key] = struct{}{}
	}
}

func (i *InitialReports) Seal(source string) {
	if i == nil {
		return
	}
	i.mu.Lock()
	defer i.mu.Unlock()
	i.sources[source] = true
}

func (i *InitialReports) Resolve(key string) {
	if i == nil {
		return
	}
	i.mu.Lock()
	defer i.mu.Unlock()
	delete(i.pending, key)
}

// PersistenceResult returns an acknowledgement only for unresolved initial work.
// Errors leave the key pending; a later successful event can still resolve it.
func (i *InitialReports) PersistenceResult(key string) func(error) {
	if i == nil {
		return nil
	}
	i.mu.Lock()
	_, pending := i.pending[key]
	i.mu.Unlock()
	if !pending {
		return nil
	}
	return func(err error) {
		if err == nil {
			i.Resolve(key)
		}
	}
}

func (i *InitialReports) Pending() []string {
	if i == nil {
		return nil
	}
	i.mu.Lock()
	defer i.mu.Unlock()
	keys := make([]string, 0, len(i.pending))
	for key := range i.pending {
		keys = append(keys, key)
	}
	return keys
}

func (i *InitialReports) Complete() bool {
	if i == nil {
		return true
	}
	i.mu.Lock()
	defer i.mu.Unlock()
	for _, sealed := range i.sources {
		if !sealed {
			return false
		}
	}
	return len(i.pending) == 0
}

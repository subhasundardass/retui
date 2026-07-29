package retui

import "sync"

type OverlayManager struct {
	mu       sync.RWMutex
	overlays map[string]bool
}

var globalOverlay = &OverlayManager{
	overlays: make(map[string]bool),
}

func OpenOverlay(id string) {
	globalOverlay.mu.Lock()
	defer globalOverlay.mu.Unlock()
	globalOverlay.overlays[id] = true
}

func CloseOverlay(id string) {
	globalOverlay.mu.Lock()
	defer globalOverlay.mu.Unlock()
	delete(globalOverlay.overlays, id)
}

func IsAnyOverlayOpen() bool {
	globalOverlay.mu.RLock()
	defer globalOverlay.mu.RUnlock()
	return len(globalOverlay.overlays) > 0
}

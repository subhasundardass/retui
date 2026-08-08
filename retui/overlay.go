package retui

import "sync"

type OverlayManager struct {
	mu       sync.RWMutex
	overlays map[string]bool
	stack    []string
}

var globalOverlay = &OverlayManager{
	overlays: make(map[string]bool),
}

func OpenOverlay(id string) {
	globalOverlay.mu.Lock()
	defer globalOverlay.mu.Unlock()
	for _, existing := range globalOverlay.stack {
		if existing == id {
			return // already open, avoid duplicate
		}
	}
	globalOverlay.stack = append(globalOverlay.stack, id)
}

func CloseOverlay(id string) {
	globalOverlay.mu.Lock()
	defer globalOverlay.mu.Unlock()
	for i, existing := range globalOverlay.stack {
		if existing == id {
			globalOverlay.stack = append(globalOverlay.stack[:i], globalOverlay.stack[i+1:]...)
			return
		}
	}
}
func TopOverlay() string {
	globalOverlay.mu.RLock()
	defer globalOverlay.mu.RUnlock()
	if len(globalOverlay.stack) == 0 {
		return ""
	}
	return globalOverlay.stack[len(globalOverlay.stack)-1]
}

func CloseTopOverlay() {
	globalOverlay.mu.Lock()
	defer globalOverlay.mu.Unlock()
	if len(globalOverlay.stack) > 0 {
		globalOverlay.stack = globalOverlay.stack[:len(globalOverlay.stack)-1]
	}
}

func IsAnyOverlayOpen() bool {
	globalOverlay.mu.RLock()
	defer globalOverlay.mu.RUnlock()
	return len(globalOverlay.stack) > 0
}

func OverlayCentered(screenW, screenH, w, h int, children ...Element) Element {
	x := (screenW - w) / 2
	y := (screenH - h) / 2
	return Overlay(x, y, children...)
}

func ResetOverlays() {
	globalOverlay.mu.Lock()
	defer globalOverlay.mu.Unlock()
	globalOverlay.stack = nil
}

package retui

import (
	"testing"
)

// ─── TestNewFocusManager ─────────────────────────────────────────────────────

func TestNewFocusManager(t *testing.T) {
	fm := NewFocusManager()
	if fm == nil {
		t.Error("NewFocusManager() returned nil")
	}
	if fm.Current() != "" {
		t.Errorf("Expected empty current focus, got %q", fm.Current())
	}
	if len(fm.order) != 0 {
		t.Errorf("Expected empty order, got %v", fm.order)
	}
	if len(fm.stack) != 0 {
		t.Errorf("Expected empty stack, got %v", fm.stack)
	}
}

// ─── TestSetAndGetFocus ──────────────────────────────────────────────────────

func TestSetAndGetFocus(t *testing.T) {
	fm := NewFocusManager()

	// Test setting focus
	fm.Focus("item1")
	if fm.Current() != "item1" {
		t.Errorf("Expected current 'item1', got %q", fm.Current())
	}
	if !fm.IsFocused("item1") {
		t.Error("Expected item1 to be focused")
	}
	if fm.IsFocused("item2") {
		t.Error("Expected item2 to not be focused")
	}

	// Test changing focus
	fm.Focus("item2")
	if fm.Current() != "item2" {
		t.Errorf("Expected current 'item2', got %q", fm.Current())
	}
	if fm.IsFocused("item1") {
		t.Error("Expected item1 to not be focused")
	}
	if !fm.IsFocused("item2") {
		t.Error("Expected item2 to be focused")
	}
}

func TestSetFocusEmptyClearsFocus(t *testing.T) {
	fm := NewFocusManager()
	fm.Focus("item1")
	fm.Focus("")

	if fm.Current() != "" {
		t.Errorf("Expected empty current focus, got %q", fm.Current())
	}
	if fm.IsFocused("item1") {
		t.Error("Expected item1 to not be focused after clearing")
	}
}

// ─── TestFocusOrder ──────────────────────────────────────────────────────────

func TestSetOrder(t *testing.T) {
	fm := NewFocusManager()
	order := []string{"a", "b", "c", "d"}
	fm.SetOrder(order)

	if len(fm.order) != len(order) {
		t.Errorf("Expected order length %d, got %d", len(order), len(fm.order))
	}
	for i, v := range order {
		if fm.order[i] != v {
			t.Errorf("Expected order[%d] = %q, got %q", i, v, fm.order[i])
		}
	}
}

func TestNextFocus(t *testing.T) {
	fm := NewFocusManager()
	fm.SetOrder([]string{"a", "b", "c"})

	// Start with no focus, Next should set to first item
	fm.Next()
	if fm.Current() != "a" {
		t.Errorf("Expected 'a', got %q", fm.Current())
	}

	// Next should cycle through order
	fm.Next()
	if fm.Current() != "b" {
		t.Errorf("Expected 'b', got %q", fm.Current())
	}
	fm.Next()
	if fm.Current() != "c" {
		t.Errorf("Expected 'c', got %q", fm.Current())
	}
	fm.Next()
	if fm.Current() != "a" {
		t.Errorf("Expected 'a' (wrap around), got %q", fm.Current())
	}
}

func TestNextFocusWithEmptyOrder(t *testing.T) {
	fm := NewFocusManager()
	fm.SetOrder([]string{})

	// Should not panic
	fm.Next()
	if fm.Current() != "" {
		t.Errorf("Expected empty current, got %q", fm.Current())
	}
}

func TestPrevFocus(t *testing.T) {
	fm := NewFocusManager()
	fm.SetOrder([]string{"a", "b", "c"})

	// Start with focus on a
	fm.Focus("a")

	fm.Prev()
	if fm.Current() != "c" {
		t.Errorf("Expected 'c' (previous from 'a'), got %q", fm.Current())
	}
	fm.Prev()
	if fm.Current() != "b" {
		t.Errorf("Expected 'b', got %q", fm.Current())
	}
	fm.Prev()
	if fm.Current() != "a" {
		t.Errorf("Expected 'a', got %q", fm.Current())
	}
}

func TestPrevFocusWithEmptyOrder(t *testing.T) {
	fm := NewFocusManager()
	fm.SetOrder([]string{})

	// Should not panic
	fm.Focus("a")
	fm.Prev()
	if fm.Current() != "a" {
		t.Errorf("Expected current to stay 'a', got %q", fm.Current())
	}
}

// ─── TestFocusOrderWithCurrentNotInOrder ────────────────────────────────────

func TestNextFocusWhenCurrentNotInOrder(t *testing.T) {
	fm := NewFocusManager()
	fm.SetOrder([]string{"a", "b", "c"})
	fm.Focus("x") // Not in order

	fm.Next()
	// Should go to first in order
	if fm.Current() != "a" {
		t.Errorf("Expected 'a' (first in order), got %q", fm.Current())
	}
}

func TestPrevFocusWhenCurrentNotInOrder(t *testing.T) {
	fm := NewFocusManager()
	fm.SetOrder([]string{"a", "b", "c"})
	fm.Focus("x") // Not in order

	fm.Prev()
	// Should wrap to last item since current not found
	if fm.Current() != "c" {
		t.Errorf("Expected 'c' (last in order), got %q", fm.Current())
	}
}

// ─── TestFocusStack ──────────────────────────────────────────────────────────

func TestPushAndPopFocus(t *testing.T) {
	fm := NewFocusManager()

	// Initial focus
	fm.Focus("main")
	if fm.Current() != "main" {
		t.Errorf("Expected 'main', got %q", fm.Current())
	}

	// Push focus (modal/dialog)
	fm.PushFocus("modal")
	if fm.Current() != "modal" {
		t.Errorf("Expected 'modal', got %q", fm.Current())
	}
	if len(fm.stack) != 1 || fm.stack[0] != "main" {
		t.Errorf("Stack should contain ['main'], got %v", fm.stack)
	}

	// Push another focus
	fm.PushFocus("submodal")
	if fm.Current() != "submodal" {
		t.Errorf("Expected 'submodal', got %q", fm.Current())
	}
	if len(fm.stack) != 2 {
		t.Errorf("Expected stack length 2, got %d", len(fm.stack))
	}
	if fm.stack[0] != "main" || fm.stack[1] != "modal" {
		t.Errorf("Stack should be ['main', 'modal'], got %v", fm.stack)
	}

	// Pop focus (close modal)
	fm.PopFocus()
	if fm.Current() != "modal" {
		t.Errorf("Expected 'modal', got %q", fm.Current())
	}
	if len(fm.stack) != 1 {
		t.Errorf("Expected stack length 1, got %d", len(fm.stack))
	}

	// Pop again
	fm.PopFocus()
	if fm.Current() != "main" {
		t.Errorf("Expected 'main', got %q", fm.Current())
	}
	if len(fm.stack) != 0 {
		t.Errorf("Expected empty stack, got %v", fm.stack)
	}
}

func TestPopFocusWithEmptyStack(t *testing.T) {
	fm := NewFocusManager()
	fm.Focus("main")

	// Pop when stack is empty should do nothing
	fm.PopFocus()
	if fm.Current() != "main" {
		t.Errorf("Expected current to stay 'main', got %q", fm.Current())
	}
	if len(fm.stack) != 0 {
		t.Errorf("Expected empty stack, got %v", fm.stack)
	}
}

func TestPushFocusWhenNoCurrentFocus(t *testing.T) {
	fm := NewFocusManager()
	// No current focus

	fm.PushFocus("modal")
	if fm.Current() != "modal" {
		t.Errorf("Expected 'modal', got %q", fm.Current())
	}
	if len(fm.stack) != 0 {
		t.Errorf("Expected empty stack when no current focus, got %v", fm.stack)
	}
}

// ─── TestGlobalFocusFunctions ────────────────────────────────────────────────

func TestGlobalFocusFunctions(t *testing.T) {
	// Reset global focus state
	globalFocus = NewFocusManager()

	// Test SetFocusOrder
	SetFocusOrder([]string{"one", "two", "three"})
	if len(globalFocus.order) != 3 {
		t.Errorf("Expected order length 3, got %d", len(globalFocus.order))
	}

	// Test SetFocus
	SetFocus("two")
	if CurrentFocus() != "two" {
		t.Errorf("Expected 'two', got %q", CurrentFocus())
	}
	if !IsFocused("two") {
		t.Error("Expected 'two' to be focused")
	}
	if IsFocused("one") {
		t.Error("Expected 'one' to not be focused")
	}

	// Test FocusNext
	FocusNext()
	if CurrentFocus() != "three" {
		t.Errorf("Expected 'three' after FocusNext, got %q", CurrentFocus())
	}

	// Test FocusPrev
	FocusPrev()
	if CurrentFocus() != "two" {
		t.Errorf("Expected 'two' after FocusPrev, got %q", CurrentFocus())
	}

	// Test Blur
	Blur()
	if CurrentFocus() != "" {
		t.Errorf("Expected empty after Blur, got %q", CurrentFocus())
	}

	// Test PushFocus and PopFocus
	SetFocus("main")
	PushFocus("modal")
	if CurrentFocus() != "modal" {
		t.Errorf("Expected 'modal', got %q", CurrentFocus())
	}
	PopFocus()
	if CurrentFocus() != "main" {
		t.Errorf("Expected 'main', got %q", CurrentFocus())
	}
}

// ─── TestFocusManagerConcurrency ─────────────────────────────────────────────

func TestFocusManagerConcurrency(t *testing.T) {
	fm := NewFocusManager()
	fm.SetOrder([]string{"a", "b", "c", "d", "e"})

	// Run concurrent operations
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			fm.Focus("a")
			fm.Next()
			fm.Prev()
			fm.PushFocus("modal")
			fm.PopFocus()
			done <- true
		}()
	}

	// Wait for all goroutines to finish
	for i := 0; i < 10; i++ {
		<-done
	}
}

// ─── TestIndexOf ─────────────────────────────────────────────────────────────

func TestIndexOf(t *testing.T) {
	fm := NewFocusManager()
	fm.SetOrder([]string{"a", "b", "c"})

	tests := []struct {
		id       string
		expected int
	}{
		{"a", 0},
		{"b", 1},
		{"c", 2},
		{"x", -1}, //Not found returns -1
	}

	for _, tt := range tests {
		result := fm.indexOf(tt.id)
		if result != tt.expected {
			t.Errorf("indexOf(%q) = %d, expected %d", tt.id, result, tt.expected)
		}
	}
}

// ─── BenchmarkTests ──────────────────────────────────────────────────────────

func BenchmarkSetFocus(b *testing.B) {
	fm := NewFocusManager()
	for i := 0; i < b.N; i++ {
		fm.Focus("item")
	}
}

func BenchmarkNextFocus(b *testing.B) {
	fm := NewFocusManager()
	fm.SetOrder([]string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"})
	fm.Focus("a")

	for i := 0; i < b.N; i++ {
		fm.Next()
	}
}

func BenchmarkPushPopFocus(b *testing.B) {
	fm := NewFocusManager()
	fm.Focus("main")

	for i := 0; i < b.N; i++ {
		fm.PushFocus("modal")
		fm.PopFocus()
	}
}

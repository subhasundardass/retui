package components

import (
	"sync"
	"testing"
	"time"

	"github.com/subhasundardass/retui/retui/window"
)

// =======================================================================
// Toast (single toast) behavior
// =======================================================================

func TestToast_Title(t *testing.T) {
	cases := []struct {
		typ  ToastType
		want string
	}{
		{ToastInfo, "INFO"},
		{ToastSuccess, "SUCCESS"},
		{ToastWarning, "WARNING"},
		{ToastError, "ERROR"},
	}
	for _, c := range cases {
		toast := &Toast{Type: c.typ}
		if got := toast.Title(); got != c.want {
			t.Errorf("Title() for type %v = %q, want %q", c.typ, got, c.want)
		}
	}
}

func TestToast_Expired(t *testing.T) {
	now := time.Now()

	persistent := &Toast{Duration: 0, Created: now.Add(-time.Hour)}
	if persistent.expired(now) {
		t.Error("toast with Duration=0 should never expire")
	}

	fresh := &Toast{Duration: time.Minute, Created: now}
	if fresh.expired(now) {
		t.Error("freshly created toast should not be expired")
	}

	stale := &Toast{Duration: time.Millisecond, Created: now.Add(-time.Hour)}
	if !stale.expired(now) {
		t.Error("toast past its duration should be expired")
	}
}

func TestToast_WidthAccountsForMessageLength(t *testing.T) {
	short := &Toast{Type: ToastInfo, Message: "hi"}
	long := &Toast{Type: ToastInfo, Message: "this is a much longer message"}

	if long.Width() <= short.Width() {
		t.Errorf("longer message should produce larger width: short=%d long=%d", short.Width(), long.Width())
	}
}

func TestToast_HeightIsFixed(t *testing.T) {
	toast := &Toast{Message: "anything"}
	if got := toast.Height(); got != 1 {
		t.Errorf("Height() = %d, want 1", got)
	}
}

// =======================================================================
// ToastManager — basic queue behavior
// =======================================================================

func TestManager_ShowAddsToast(t *testing.T) {
	m := NewToastManager()

	id := m.Show("hello world")
	if id == 0 {
		t.Fatal("Show() returned zero ID")
	}

	toasts := m.Toasts()
	if len(toasts) != 1 {
		t.Fatalf("expected 1 toast, got %d", len(toasts))
	}
	if toasts[0].Message != "hello world" {
		t.Errorf("Message = %q, want %q", toasts[0].Message, "hello world")
	}
	if toasts[0].ID != id {
		t.Errorf("stored toast ID = %d, want %d", toasts[0].ID, id)
	}
}

func TestManager_IDsAreUniqueAndIncreasing(t *testing.T) {
	m := NewToastManager()

	id1 := m.Show("first")
	id2 := m.Show("second")

	if id1 == id2 {
		t.Fatal("expected unique IDs for separate toasts")
	}
	if id2 <= id1 {
		t.Errorf("expected increasing IDs, got id1=%d id2=%d", id1, id2)
	}
}

func TestManager_DefaultsAppliedWhenNoOptionsGiven(t *testing.T) {
	m := NewToastManager() // defaults: Info, 3s, BottomLeft

	m.Show("plain")
	toasts := m.Toasts()
	if len(toasts) != 1 {
		t.Fatal("expected 1 toast")
	}
	got := toasts[0]
	if got.Type != ToastInfo {
		t.Errorf("Type = %v, want ToastInfo", got.Type)
	}
	if got.Position != ToastBottomLeft {
		t.Errorf("Position = %v, want ToastBottomLeft", got.Position)
	}
	if got.Duration != 3*time.Second {
		t.Errorf("Duration = %v, want 3s", got.Duration)
	}
}

func TestManager_PerCallOptionsOverrideDefaults(t *testing.T) {
	m := NewToastManager()

	m.Show("custom",
		WithType(ToastError),
		WithPosition(ToastTopRight),
		WithDuration(10*time.Second),
	)

	got := m.Toasts()[0]
	if got.Type != ToastError {
		t.Errorf("Type = %v, want ToastError", got.Type)
	}
	if got.Position != ToastTopRight {
		t.Errorf("Position = %v, want ToastTopRight", got.Position)
	}
	if got.Duration != 10*time.Second {
		t.Errorf("Duration = %v, want 10s", got.Duration)
	}
}

// =======================================================================
// ToastManager — Configure (manager-level defaults)
// =======================================================================

func TestManager_ConfigureChangesDefaults(t *testing.T) {
	m := NewToastManager()

	m.Configure(
		WithDefaultPosition(ToastTopLeft),
		WithDefaultDuration(7*time.Second),
	)

	m.Show("after configure")
	got := m.Toasts()[0]

	if got.Position != ToastTopLeft {
		t.Errorf("Position = %v, want ToastTopLeft", got.Position)
	}
	if got.Duration != 7*time.Second {
		t.Errorf("Duration = %v, want 7s", got.Duration)
	}
}

func TestManager_ConfigureThenPerCallOverrideStillWins(t *testing.T) {
	m := NewToastManager()
	m.Configure(WithDefaultPosition(ToastTopLeft))

	m.Show("explicit wins", WithPosition(ToastBottomRight))
	got := m.Toasts()[0]

	if got.Position != ToastBottomRight {
		t.Errorf("Position = %v, want ToastBottomRight (per-call override)", got.Position)
	}
}

// =======================================================================
// ToastManager — type-specific helpers (ShowSuccess/Warning/Error)
// =======================================================================

func TestShowHelpers_SetCorrectType(t *testing.T) {
	ClearToasts()
	t.Cleanup(ClearToasts)

	ShowSuccess("ok")
	ShowWarning("careful")
	ShowError("broken")

	toasts := defaultToasts.Toasts()
	if len(toasts) != 3 {
		t.Fatalf("expected 3 toasts, got %d", len(toasts))
	}

	want := map[string]ToastType{
		"ok":      ToastSuccess,
		"careful": ToastWarning,
		"broken":  ToastError,
	}
	for _, toast := range toasts {
		wantType, ok := want[toast.Message]
		if !ok {
			t.Fatalf("unexpected toast message %q", toast.Message)
		}
		if toast.Type != wantType {
			t.Errorf("message %q: Type = %v, want %v", toast.Message, toast.Type, wantType)
		}
	}
}

func TestShowError_AcceptsAdditionalOptions(t *testing.T) {
	ClearToasts()
	t.Cleanup(ClearToasts)

	ShowError("positioned error", WithPosition(ToastTopRight))

	toasts := defaultToasts.Toasts()
	if len(toasts) != 1 {
		t.Fatalf("expected 1 toast, got %d", len(toasts))
	}
	if toasts[0].Type != ToastError {
		t.Errorf("Type = %v, want ToastError", toasts[0].Type)
	}
	if toasts[0].Position != ToastTopRight {
		t.Errorf("Position = %v, want ToastTopRight", toasts[0].Position)
	}
}

// =======================================================================
// ToastManager — Dismiss / Clear
// =======================================================================

func TestManager_DismissRemovesOnlyMatchingToast(t *testing.T) {
	m := NewToastManager()

	id1 := m.Show("keep me")
	id2 := m.Show("remove me")

	m.Dismiss(id2)

	toasts := m.Toasts()
	if len(toasts) != 1 {
		t.Fatalf("expected 1 toast after dismiss, got %d", len(toasts))
	}
	if toasts[0].ID != id1 {
		t.Errorf("remaining toast ID = %d, want %d", toasts[0].ID, id1)
	}
}

func TestManager_DismissUnknownIDIsNoop(t *testing.T) {
	m := NewToastManager()
	m.Show("still here")

	m.Dismiss(99999) // never existed

	if len(m.Toasts()) != 1 {
		t.Error("dismissing unknown ID should not affect existing toasts")
	}
}

func TestManager_ClearRemovesEverything(t *testing.T) {
	m := NewToastManager()
	m.Show("one")
	m.Show("two")
	m.Show("three")

	m.Clear()

	if len(m.Toasts()) != 0 {
		t.Error("expected empty queue after Clear()")
	}
}

// =======================================================================
// ToastManager — expiry / sweeping
// =======================================================================

func TestManager_ExpiredToastsAreSweptOnRead(t *testing.T) {
	m := NewToastManager()
	m.Show("blink and it's gone", WithDuration(10*time.Millisecond))

	if !m.HasToast() {
		t.Fatal("toast should be visible immediately after Show()")
	}

	time.Sleep(30 * time.Millisecond)

	if m.HasToast() {
		t.Error("expected expired toast to be swept and HasToast() to return false")
	}
	if len(m.Toasts()) != 0 {
		t.Error("expected Toasts() to return empty slice after expiry")
	}
}

func TestManager_ZeroDurationPersists(t *testing.T) {
	m := NewToastManager()
	m.Show("forever", WithDuration(0))

	time.Sleep(20 * time.Millisecond)

	if !m.HasToast() {
		t.Error("toast with Duration=0 should never expire")
	}
}

func TestManager_MixedExpiryOnlyRemovesExpiredOnes(t *testing.T) {
	m := NewToastManager()
	shortID := m.Show("short", WithDuration(10*time.Millisecond))
	longID := m.Show("long", WithDuration(time.Hour))

	time.Sleep(30 * time.Millisecond)

	toasts := m.Toasts()
	if len(toasts) != 1 {
		t.Fatalf("expected 1 surviving toast, got %d", len(toasts))
	}
	if toasts[0].ID != longID {
		t.Errorf("surviving toast ID = %d, want %d (long-lived)", toasts[0].ID, longID)
	}
	_ = shortID
}

// =======================================================================
// ToastManager — OnChange notifications
// =======================================================================

func TestManager_OnChangeFiresOnShow(t *testing.T) {
	m := NewToastManager()

	var mu sync.Mutex
	fired := false
	m.OnChange(func() {
		mu.Lock()
		fired = true
		mu.Unlock()
	})

	m.Show("trigger notify")

	mu.Lock()
	defer mu.Unlock()
	if !fired {
		t.Error("expected OnChange callback to fire on Show()")
	}
}

func TestManager_OnChangeFiresOnDismissAndClear(t *testing.T) {
	m := NewToastManager()

	var mu sync.Mutex
	count := 0
	m.OnChange(func() {
		mu.Lock()
		count++
		mu.Unlock()
	})

	id := m.Show("a") // 1
	m.Dismiss(id)     // 2
	m.Show("b")       // 3
	m.Clear()         // 4

	mu.Lock()
	defer mu.Unlock()
	if count != 4 {
		t.Errorf("OnChange fired %d times, want 4", count)
	}
}

func TestManager_OnChangeFiresOnExpirySweep(t *testing.T) {
	m := NewToastManager()

	done := make(chan struct{}, 1)
	m.OnChange(func() {
		select {
		case done <- struct{}{}:
		default:
		}
	})

	m.Show("expires soon", WithDuration(10*time.Millisecond))
	// drain the Show() notification
	<-done

	select {
	case <-done:
		// notified again via the AfterFunc sweep — good
	case <-time.After(200 * time.Millisecond):
		t.Error("expected OnChange to fire again when the toast expired")
	}
}

// =======================================================================
// ToastManager — concurrency (run with -race)
// =======================================================================

func TestManager_ConcurrentShowIsSafe(t *testing.T) {
	m := NewToastManager()

	var wg sync.WaitGroup
	const goroutines = 50
	const perGoroutine = 20

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			for j := 0; j < perGoroutine; j++ {
				m.Show("concurrent", WithDuration(time.Hour))
			}
		}(i)
	}
	wg.Wait()

	if got := len(m.Toasts()); got != goroutines*perGoroutine {
		t.Errorf("expected %d toasts, got %d", goroutines*perGoroutine, got)
	}
}

func TestManager_ConcurrentReadWriteIsSafe(t *testing.T) {
	m := NewToastManager()
	m.Show("seed", WithDuration(time.Hour))

	var wg sync.WaitGroup
	stop := make(chan struct{})

	// Writers
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stop:
					return
				default:
					m.Show("x", WithDuration(5*time.Millisecond))
				}
			}
		}()
	}

	// Readers (HasToast internally sweeps — exercises the write-under-lock path)
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stop:
					return
				default:
					m.HasToast()
					m.Toasts()
				}
			}
		}()
	}

	time.Sleep(50 * time.Millisecond)
	close(stop)
	wg.Wait()
}

// =======================================================================
// ToastLayer
// =======================================================================

func TestToastLayer_EmptyQueueReturnsNil(t *testing.T) {
	m := NewToastManager()
	if overlays := ToastLayer(m); overlays != nil {
		t.Errorf("expected nil overlays for empty queue, got %d elements", len(overlays))
	}
}

func TestToastLayer_ReturnsOneOverlayPerToast(t *testing.T) {
	window.SetScreenSize(100, 40) // ensure deterministic screen size for this test

	m := NewToastManager()
	m.Show("one", WithDuration(time.Hour))
	m.Show("two", WithDuration(time.Hour))
	m.Show("three", WithDuration(time.Hour))

	overlays := ToastLayer(m)
	if len(overlays) != 3 {
		t.Errorf("expected 3 overlay elements, got %d", len(overlays))
	}
}

func TestToastLayer_NilManagerFallsBackToDefault(t *testing.T) {
	ClearToasts()
	t.Cleanup(ClearToasts)

	ShowToast("via default manager", WithDuration(time.Hour))

	overlays := ToastLayer(nil)
	if len(overlays) != 1 {
		t.Errorf("expected 1 overlay via default manager, got %d", len(overlays))
	}
}

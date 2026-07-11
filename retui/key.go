package retui

import "bytes"

// KeyCode identifies special (non-printable) keys.
type KeyCode int

var Keys = make(chan Key, 4)

// CurrentKey holds the key being processed in the current render pass.
var CurrentKey Key

const (
	KeyNone KeyCode = iota

	// ─── Function Keys ──────────────────────────────────────────────────────
	KeyF1
	KeyF2
	KeyF3
	KeyF4
	KeyF5
	KeyF6
	KeyF7
	KeyF8
	KeyF9
	KeyF10
	KeyF11
	KeyF12

	// ─── Navigation Keys ──────────────────────────────────────────────────
	KeyUp
	KeyDown
	KeyLeft
	KeyRight
	KeyHome
	KeyEnd
	KeyPageUp
	KeyPageDown
	KeyInsert
	KeyDelete

	// ─── Special Keys ──────────────────────────────────────────────────────
	KeyEnter
	KeyBackspace
	KeyEscape
	KeyTab
	KeyShiftTab
	KeySpace
	KeyCtrlC
	KeyPaste

	// ─── Ctrl+Letter Keys ──────────────────────────────────────────────────
	KeyCtrlA
	KeyCtrlB
	KeyCtrlD
	KeyCtrlE
	KeyCtrlF
	KeyCtrlG
	KeyCtrlH // 0x08 - same as Backspace
	KeyCtrlI // 0x09 - same as Tab
	KeyCtrlJ // 0x0A - same as Enter
	KeyCtrlK
	KeyCtrlL
	KeyCtrlM // 0x0D - same as Enter
	KeyCtrlN
	KeyCtrlO
	KeyCtrlP
	KeyCtrlQ
	KeyCtrlR
	KeyCtrlS
	KeyCtrlT
	KeyCtrlU
	KeyCtrlV
	KeyCtrlW
	KeyCtrlX
	KeyCtrlY
	KeyCtrlZ

	// ─── Alt+Letter Keys ──────────────────────────────────────────────────
	KeyAltA
	KeyAltB
	KeyAltC
	KeyAltD
	KeyAltE
	KeyAltF
	KeyAltG
	KeyAltH
	KeyAltI
	KeyAltJ
	KeyAltK
	KeyAltL
	KeyAltM
	KeyAltN
	KeyAltO
	KeyAltP
	KeyAltQ
	KeyAltR
	KeyAltS
	KeyAltT
	KeyAltU
	KeyAltV
	KeyAltW
	KeyAltX
	KeyAltY
	KeyAltZ
)

// Key represents a single keyboard event. Either Code or Rune is set.
// When Code == KeyPaste, Paste holds the full pasted text.
type Key struct {
	Code  KeyCode
	Rune  rune
	Paste string
}

// ─── Parse Functions ────────────────────────────────────────────────────────

// ParseKey converts raw terminal bytes into a Key.
func ParseKey(b []byte) Key {
	if len(b) == 0 {
		return Key{}
	}

	// ─── Escape Sequences (CSI) ────────────────────────────────────────────

	if len(b) >= 3 && b[0] == 0x1B && b[1] == '[' {
		switch b[2] {
		// Arrow keys
		case 'A':
			return Key{Code: KeyUp}
		case 'B':
			return Key{Code: KeyDown}
		case 'C':
			return Key{Code: KeyRight}
		case 'D':
			return Key{Code: KeyLeft}
		case 'Z':
			return Key{Code: KeyShiftTab}

		// Function keys (F1-F4)
		case 'P':
			return Key{Code: KeyF1}
		case 'Q':
			return Key{Code: KeyF2}
		case 'R':
			return Key{Code: KeyF3}
		case 'S':
			return Key{Code: KeyF4}

		// Home, End
		case 'H':
			return Key{Code: KeyHome}
		case 'F':
			return Key{Code: KeyEnd}

		// Insert, Delete, PageUp, PageDown
		case '2', '3', '5', '6':
			if len(b) >= 4 && b[3] == '~' {
				switch b[2] {
				case '2':
					return Key{Code: KeyInsert}
				case '3':
					return Key{Code: KeyDelete}
				case '5':
					return Key{Code: KeyPageUp}
				case '6':
					return Key{Code: KeyPageDown}
				}
			}
		}
	}

	// ─── Function Keys (F5-F12) with longer sequences ────────────────────
	if len(b) >= 3 && b[0] == 0x1B && b[1] == 'O' {
		switch b[2] {
		case 'P':
			return Key{Code: KeyF1}
		case 'Q':
			return Key{Code: KeyF2}
		case 'R':
			return Key{Code: KeyF3}
		case 'S':
			return Key{Code: KeyF4}
		}
	}
	if len(b) >= 5 && b[0] == 0x1B && b[1] == '[' && b[2] == '1' {
		switch b[3] {
		case '5':
			if b[4] == '~' {
				return Key{Code: KeyF5}
			}
		case '7':
			if b[4] == '~' {
				return Key{Code: KeyF6}
			}
		case '8':
			if b[4] == '~' {
				return Key{Code: KeyF7}
			}
		case '9':
			if b[4] == '~' {
				return Key{Code: KeyF8}
			}
		}
	}

	if len(b) >= 5 && b[0] == 0x1B && b[1] == '[' && b[2] == '2' {
		switch b[3] {
		case '0':
			if b[4] == '~' {
				return Key{Code: KeyF9}
			}
		case '1':
			if b[4] == '~' {
				return Key{Code: KeyF10}
			}
		case '3':
			if b[4] == '~' {
				return Key{Code: KeyF11}
			}
		case '4':
			if b[4] == '~' {
				return Key{Code: KeyF12}
			}
		}
	}

	// ─── Alt Keys (Alt+Letter) ────────────────────────────────────────────

	if len(b) >= 2 && b[0] == 0x1B {
		if b[1] >= 0x61 && b[1] <= 0x7A {
			return Key{Code: KeyCode(int(KeyAltA) + int(b[1]-0x61))}
		}
		if b[1] >= 0x41 && b[1] <= 0x5A {
			return Key{Code: KeyCode(int(KeyAltA) + int(b[1]-0x41))}
		}
	}

	// ─── Single Byte Keys ──────────────────────────────────────────────────

	switch b[0] {
	case 0x1B:
		return Key{Code: KeyEscape}
	case 0x0D, 0x0A:
		return Key{Code: KeyEnter}
	case 0x7F, 0x08:
		return Key{Code: KeyBackspace}
	case 0x09:
		return Key{Code: KeyTab}
	case 0x20:
		return Key{Code: KeySpace}
	case 0x03:
		return Key{Code: KeyCtrlC}

	// ─── Ctrl+Letter (0x01 = Ctrl+A, 0x1A = Ctrl+Z) ─────────────────────
	// Note: 0x08 (Ctrl+H), 0x09 (Ctrl+I), 0x0A (Ctrl+J), 0x0D (Ctrl+M)
	// are already handled above as Backspace, Tab, Enter, Enter

	case 0x01:
		return Key{Code: KeyCtrlA}
	case 0x02:
		return Key{Code: KeyCtrlB}
	case 0x04:
		return Key{Code: KeyCtrlD}
	case 0x05:
		return Key{Code: KeyCtrlE}
	case 0x06:
		return Key{Code: KeyCtrlF}
	case 0x07:
		return Key{Code: KeyCtrlG}
	// 0x08 is Backspace (Ctrl+H) - already handled
	// 0x09 is Tab (Ctrl+I) - already handled
	// 0x0A is Enter (Ctrl+J) - already handled
	case 0x0B:
		return Key{Code: KeyCtrlK}
	case 0x0C:
		return Key{Code: KeyCtrlL}
	// 0x0D is Enter (Ctrl+M) - already handled
	case 0x0E:
		return Key{Code: KeyCtrlN}
	case 0x0F:
		return Key{Code: KeyCtrlO}
	case 0x10:
		return Key{Code: KeyCtrlP}
	case 0x11:
		return Key{Code: KeyCtrlQ}
	case 0x12:
		return Key{Code: KeyCtrlR}
	case 0x13:
		return Key{Code: KeyCtrlS}
	case 0x14:
		return Key{Code: KeyCtrlT}
	case 0x15:
		return Key{Code: KeyCtrlU}
	case 0x16:
		return Key{Code: KeyCtrlV}
	case 0x17:
		return Key{Code: KeyCtrlW}
	case 0x18:
		return Key{Code: KeyCtrlX}
	case 0x19:
		return Key{Code: KeyCtrlY}
	case 0x1A:
		return Key{Code: KeyCtrlZ}
	}

	// ─── Printable Characters ─────────────────────────────────────────────

	return Key{Rune: rune(b[0])}
}

// ─── KeyScanner ─────────────────────────────────────────────────────────────

var (
	pasteStart = []byte{0x1B, '[', '2', '0', '0', '~'}
	pasteEnd   = []byte{0x1B, '[', '2', '0', '1', '~'}
)

// KeyScanner is a stateful parser that converts raw stdin reads into Key events.
type KeyScanner struct {
	inPaste  bool
	pasteBuf []byte
}

// Feed consumes one chunk of stdin bytes and returns any complete Key events.
func (s *KeyScanner) Feed(b []byte) []Key {
	var keys []Key
	for len(b) > 0 {
		if s.inPaste {
			s.pasteBuf = append(s.pasteBuf, b...)
			b = nil
			idx := bytes.Index(s.pasteBuf, pasteEnd)
			if idx < 0 {
				return keys
			}
			pasted := string(s.pasteBuf[:idx])
			rest := append([]byte(nil), s.pasteBuf[idx+len(pasteEnd):]...)
			s.pasteBuf = nil
			s.inPaste = false
			keys = append(keys, Key{Code: KeyPaste, Paste: pasted})
			b = rest
			continue
		}
		if bytes.HasPrefix(b, pasteStart) {
			b = b[len(pasteStart):]
			s.inPaste = true
			continue
		}
		consumed := 1
		if b[0] == 0x1B && len(b) >= 3 && (b[1] == '[' || b[1] == 'O') {
			consumed = 3
		}
		keys = append(keys, ParseKey(b[:consumed]))
		b = b[consumed:]
	}
	return keys
}

// ─── Helper Functions ──────────────────────────────────────────────────────

// IsPrintable returns true if the key is a printable character
func (k Key) IsPrintable() bool {
	return k.Code == KeyNone && k.Rune >= 32 && k.Rune <= 126
}

// IsModifier returns true if the key is a modifier combination
func (k Key) IsModifier() bool {
	return k.Code >= KeyCtrlA && k.Code <= KeyCtrlZ
}

// IsFunctionKey returns true if the key is F1-F12
func (k Key) IsFunctionKey() bool {
	return k.Code >= KeyF1 && k.Code <= KeyF12
}

// IsNavigationKey returns true for arrow keys, Home, End, PageUp, PageDown
func (k Key) IsNavigationKey() bool {
	switch k.Code {
	case KeyUp, KeyDown, KeyLeft, KeyRight, KeyHome, KeyEnd, KeyPageUp, KeyPageDown:
		return true
	default:
		return false
	}
}

// String returns a readable name for the key
func (k Key) String() string {
	if k.Rune != 0 {
		return string(k.Rune)
	}
	switch k.Code {
	case KeyNone:
		return "None"
	case KeyEnter:
		return "Enter"
	case KeyBackspace:
		return "Backspace"
	case KeyEscape:
		return "Escape"
	case KeyTab:
		return "Tab"
	case KeyShiftTab:
		return "ShiftTab"
	case KeyUp:
		return "Up"
	case KeyDown:
		return "Down"
	case KeyLeft:
		return "Left"
	case KeyRight:
		return "Right"
	case KeySpace:
		return "Space"
	case KeyCtrlC:
		return "CtrlC"
	case KeyPaste:
		return "Paste"
	case KeyHome:
		return "Home"
	case KeyEnd:
		return "End"
	case KeyPageUp:
		return "PageUp"
	case KeyPageDown:
		return "PageDown"
	case KeyInsert:
		return "Insert"
	case KeyDelete:
		return "Delete"
	case KeyF1:
		return "F1"
	case KeyF2:
		return "F2"
	case KeyF3:
		return "F3"
	case KeyF4:
		return "F4"
	case KeyF5:
		return "F5"
	case KeyF6:
		return "F6"
	case KeyF7:
		return "F7"
	case KeyF8:
		return "F8"
	case KeyF9:
		return "F9"
	case KeyF10:
		return "F10"
	case KeyF11:
		return "F11"
	case KeyF12:
		return "F12"
	default:
		if k.Code >= KeyCtrlA && k.Code <= KeyCtrlZ {
			return "Ctrl" + string('A'+rune(k.Code-KeyCtrlA))
		}
		if k.Code >= KeyAltA && k.Code <= KeyAltZ {
			return "Alt" + string('A'+rune(k.Code-KeyAltA))
		}
		return "Unknown"
	}
}

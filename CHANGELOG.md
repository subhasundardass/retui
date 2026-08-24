# Changelog

All notable changes to this project will be documented in this file.

The format is based on **Keep a Changelog**, and this project follows **Semantic Versioning** where practical.

---

## [v0.9.0] - 2026-08-22

### Added

- Debugger Option (F12) — Added an F12 keyboard shortcut to toggle the debugger.
- Element Overflow & Scrolling — Added overflow, scrollX, scrollY, and wrap support to Element.
- Batch Rendering Support for Form — Form setters now participate correctly in retui.Batch(). Multiple SetField calls inside a batch are coalesced into a single redraw, matching the rendering behavior of UseState.

### Updates

- Form state updates now correctly trigger rendering while respecting Batch() boundaries.

### Updates

- Update cached screenWidth/screenHeight
- Bug Fixed in Window Center
- Bug fixed Panel Children Height

- ***

## [v0.8.0] - 2026-08-07

### Added

- Form
- Disable, Hidden & Readonly to all Input

### Updates

- syscall.SIGWINCH (Windows/Linux)

-

## [v0.7.0] - 2026-08-03

### Added

- Add Border Title with Style
- Added Toast component
- Added UseRef, UseMemo, UseReducer Hooks in State

### Updates

- Update Responsive UI

---

## [v0.6.0] - 2026-07-31

### Added

- Added Margin to Box
- Added All component Test
- Added Security Check Test

### Updates

- Update: Render(), Layout()

---

## [v0.5.0] - 2026-07-22

### Added

- Added Percent sizing mode
- Added overflow clamping
- Added ContentBuilder in Element
- Added resolveDeferred() + hasDeferred() in renderer, called at the top of Render()
- Select Dropdown with Filter (Static and Api)
- Added explicitWidth/explicitHeight bools on tableConfig

### Updates

- Changed: Render() no longer calls t.build() directly

---

## [v0.4.0] - 2026-07-18

### Added

- New Table component.
- New Select Picker component.
- Added Title Element in Panel Component
- Test file added

### Updates

- Update UseState() - Verify the type before cast.
- Update Navigation. Now PushScreen will accept params
- Propagation control - Stop keys from bubbling up
- Global handlers conflict - Multiple handlers fighting
- Update Window Documentation (https://github.com/subhasundardass/retui/wiki/Window-System)

---

## [v0.3.0] - 2026-07-14

### Added

- New Badge component.
- New Spinner component.
- New Progress component.

### Changed

- Improved the state system for better flexibility and consistency

---

## [v0.2.0] - 2026-07-13

### Added

- New Panel component.
- New List component.
- New Date Input component.

### Changed

- Updated the layout system for better flexibility and consistency
- Improved the rendering engine for faster terminal updates.

### Fixed

- Fixed several bugs and improved overall stability.

---

## [v0.1.0] - 2026-07-11

### Added

- Initial public release of Retui.
- Window management system.
- Modal window support.
- Flexible layout engine.
- Basic widgets.
- Focus management.
- Keyboard navigation.
- Event-driven architecture.
- Pure Go implementation.
- Initial documentation.

### Changed

- Improved rendering performance.
- Refined layout calculations.

### Fixed

- Minor rendering issues.
- Focus handling edge cases.

---

## Notes

Future releases will continue documenting:

- Added features
- Changed behavior
- Deprecated functionality
- Removed features
- Fixed bugs
- Security improvements

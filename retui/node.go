package retui

// ============================================================================
// Element Tree — Node Definitions
// ============================================================================

// This file defines the core types for building the element tree that
// represents the UI structure. Elements are nodes in a tree; each node can
// have children, layout properties, styling, and content (text/markdown).
//
// Element Tree Basics:
//   - Each Element is a node with optional children and content
//   - ElementType determines which fields are active and how node behaves
//   - Layout via LayoutProps (uses layout.go sizing and positioning)
//   - Styling via Style (colors, borders, fonts, etc.)
//   - Hooks via Props.Values map (custom application state)
//
// Building Elements:
//   Use constructors (NewBox, NewText, etc.) or set fields manually.
//   Prefer constructors for safety and clarity.
//
// Example:
//   root := Element{
//       Type:   ElementBox,
//       Layout: LayoutProps{Direction: Row},
//       Children: []Element{
//           {Type: ElementText, Text: "Hello"},
//           {Type: ElementText, Text: "World"},
//       },
//   }

// ============================================================================
// MarkdownContent — Pre-Parsed Markdown
// ============================================================================

// MarkdownContent carries pre-parsed markdown with per-character styling.
//
// The Lines field contains rendered output: each line is a sequence of
// styled cells (character + color/formatting).
//
// Usage:
//   - Lines is exported for reading by render engine and tests
//   - Do NOT modify Lines directly; instead set Element.MarkdownText
//   - To change markdown: set MarkdownText, let render re-parse it
//   - MarkdownContent is updated automatically during render
//
// Note: markdownLine is private (intentional). Markdown parsing is internal
// to render.go; external code should not construct lines manually. This
// ensures rendering consistency and simplifies the API.
//
// Example:
//
//	elem := Element{
//	    Type:         ElementMarkdown,
//	    MarkdownText: "# Title\n\nParagraph",  // Set raw markdown
//	    // Markdown will be populated by render engine
//	}
type MarkdownContent struct {
	// Lines is the rendered output; read-only from outside render engine.
	// Each line is a sequence of styled cells (character + styling).
	// Updated by render engine when MarkdownText is set.
	Lines []markdownLine
}

// ============================================================================
// ElementType — Element Category
// ============================================================================

// ElementType identifies the category and behavior of an Element.
//
// Each ElementType determines:
//   - Which Element fields are actually used
//   - How the element is laid out and positioned
//   - How the element is rendered to the screen
//   - Which sizing and alignment rules apply
//
// See Element documentation for detailed field usage per type.
type ElementType int

const (
	// ElementBox is a container for laying out children.
	//
	// Usage:
	//   - Type: ElementBox
	//   - Uses: Children, Layout, Style
	//   - Ignores: Text, Markdown, MarkdownText, Wrap, ContentBuilder, etc.
	//
	// Behavior:
	//   - Renders background fill and optional border
	//   - Arranges children via Layout (direction, gap, alignment, justify)
	//   - Children sized/positioned according to Layout properties
	//   - Supports responsive sizing (Grow, Fit, Percent)
	//
	// Example:
	//   Element{
	//       Type: ElementBox,
	//       Layout: LayoutProps{
	//           Direction: Row,
	//           WidthSizing: Grow(1),
	//           HeightSizing: Fit(),
	//       },
	//       Children: []Element{...},
	//   }
	ElementBox ElementType = iota

	// ElementText is a single line of unformatted text.
	//
	// Usage:
	//   - Type: ElementText
	//   - Uses: Text, Style
	//   - Ignores: Children, Markdown, Wrap, Layout, ContentBuilder, etc.
	//
	// Behavior:
	//   - Renders text on a single line
	//   - Width determined by text length (rune count with RuneWidth)
	//   - Height is always 1 row
	//   - No wrapping or line breaking
	//   - Text exceeding container width is clipped
	//
	// Example:
	//   Element{
	//       Type: ElementText,
	//       Text: "Status: OK",
	//       Style: Style{foreground: ColorGreen},
	//   }
	ElementText

	// ElementMultilineText is text that may span multiple lines.
	//
	// Usage:
	//   - Type: ElementMultilineText
	//   - Uses: Text, Wrap, Style, Layout
	//   - Ignores: Children, Markdown, ContentBuilder, etc.
	//
	// Behavior:
	//   - If Wrap=true: Text wraps at container width using reflow callback
	//   - If Wrap=false: Text split on newlines (\n); no word wrapping
	//   - Wrapping respects word boundaries (space-separated)
	//   - Words longer than width are hard-broken at width boundary
	//   - Height expands to fit all lines
	//
	// Reflow Callback:
	//   - When Wrap=true, layout engine calls reflow callback with allocated width
	//   - Callback computes wrapped height; updates Element.Layout
	//   - Second layout pass propagates height to ancestor containers
	//   - Allows containers around wrapped text to grow to fit content
	//
	// Example (Wrapped):
	//   Element{
	//       Type: ElementMultilineText,
	//       Text: "Lorem ipsum dolor sit amet...",
	//       Wrap: true,
	//       Layout: LayoutProps{WidthSizing: Grow(1)},
	//   }
	//
	// Example (Raw):
	//   Element{
	//       Type: ElementMultilineText,
	//       Text: "Line 1\nLine 2\nLine 3",
	//       Wrap: false,
	//   }
	ElementMultilineText

	// ElementMarkdown is rendered markdown with per-character styling.
	//
	// Usage:
	//   - Type: ElementMarkdown
	//   - Uses: MarkdownText (source), Markdown (rendered), Style, Layout
	//   - Ignores: Text, Wrap, Children, ContentBuilder, etc.
	//
	// Behavior:
	//   - MarkdownText: Raw markdown source (re-parsed on resize if non-empty)
	//   - Markdown: Pre-parsed output (used if MarkdownText is empty)
	//   - Renders multi-line text with inline styling (bold, italic, colors)
	//   - Each character has its own style (color, formatting)
	//   - Automatically wraps at container width during render
	//   - Height expands to fit rendered markdown
	//
	// Reflow:
	//   - Similar to wrapped text: reflow callback computes height from width
	//   - Allows containers around markdown to grow to fit content
	//
	// Example:
	//   Element{
	//       Type: ElementMarkdown,
	//       MarkdownText: "# Title\n\nParagraph with **bold** and *italic*",
	//       Layout: LayoutProps{WidthSizing: Grow(1)},
	//   }
	ElementMarkdown

	// ElementOverlay is positioned at absolute screen coordinates.
	//
	// Usage:
	//   - Type: ElementOverlay
	//   - Uses: OverlayX, OverlayY, Children, Style
	//   - Ignores: Text, Markdown, Wrap, Layout, ContentBuilder, etc.
	//
	// Behavior:
	//   - Children are positioned at absolute coordinates (OverlayX, OverlayY)
	//   - Ignores normal flow layout entirely
	//   - Zero-size in flow (doesn't push siblings or expand parents)
	//   - Painted on top of main tree (deferred to final render pass)
	//   - Used for modals, tooltips, dropdowns
	//   - Negative coordinates allowed (off-screen overlays render nothing)
	//
	// Positioning:
	//   - OverlayX: Column (left edge)
	//   - OverlayY: Row (top edge)
	//   - Both are screen-relative (not parent-relative)
	//
	// Example (Modal):
	//   Element{
	//       Type: ElementOverlay,
	//       OverlayX: 10,
	//       OverlayY: 5,
	//       Children: []Element{
	//           {Type: ElementBox, Text: "Modal content"},
	//       },
	//   }
	ElementOverlay
)

// ============================================================================
// Props — Application Properties & Hooks
// ============================================================================

// Props carries layout configuration and generic hook properties.
//
// Two Purposes:
//  1. Layout properties (Direction, Gap, Padding, Align, Justify, Width, Height)
//     These mirror LayoutProps but with a different API (Props uses Width/Height,
//     LayoutProps uses WidthSizing/HeightSizing). Kept for backward compatibility.
//  2. Generic hook data (Values map)
//     Custom application state accessible to hooks via Props.Get()
//
// Sizing Defaults:
//   - DO NOT rely on Props width/height defaults
//   - Always use NewProps() to get correct defaults
//   - Zero value Props produces Fixed(0) width/height (rarely useful)
//   - NewProps() sets Width: Fit(), Height: Fit() (sensible defaults)
//
// Relationship with LayoutProps:
//   - Element.Layout (LayoutProps) is the authoritative layout configuration
//   - It mirrors layout.go exactly and is what the layout algorithm uses
//   - Props.Width/Height are shadows for backward compatibility
//   - New code should set Element.Layout, not Element.Props
//   - Props.Values is useful for custom hook state
//
// Example (Correct Usage):
//
//	props := NewProps()
//	props.Values["counter"] = 42
//
// Example (Legacy Usage — still works):
//
//	props := Props{Width: Grow(1), Height: Fit(), Values: make(map[string]any)}
type Props struct {
	// Direction is the primary layout axis (Row or Column).
	Direction Direction
	// Gap is the minimum spacing between adjacent children on main axis.
	Gap int
	// Padding reserves space inside the element [top, right, bottom, left].
	Padding [4]int
	// Margin reserves space outside the element [top, right, bottom, left].
	Margin [4]int
	// Align controls child alignment on the cross axis.
	Align Alignment
	// Justify controls space distribution along the main axis.
	Justify Justify
	// Width sizing for this element (legacy; prefer Element.Layout.WidthSizing).
	// Zero value is Fixed(0); use NewProps() for sensible defaults.
	Width Sizing
	// Height sizing for this element (legacy; prefer Element.Layout.HeightSizing).
	// Zero value is Fixed(0); use NewProps() for sensible defaults.
	Height Sizing
	// Values is a map for custom hook properties (application state).
	// Accessible via Props.Get(key) or Values[key] directly.
	Values map[string]any
}

// Get returns the value associated with key in Props.Values, or nil if not found.
func (p Props) Get(key string) any {
	if p.Values == nil {
		return nil
	}
	return p.Values[key]
}

// NewProps creates a Props with sensible defaults.
//
// Defaults:
//   - Width: Fit() (shrink-wrap to children)
//   - Height: Fit() (shrink-wrap to children)
//   - Values: empty map
//   - Direction, Gap, Padding, Align, Justify: zero values
//
// Use NewProps() instead of Props{} to avoid accidental Fixed(0) sizing.
//
// Example:
//
//	props := NewProps()
//	props.Width = Grow(1)   // Fill parent width
//	props.Values["state"] = myState
func NewProps() Props {
	return Props{
		Width:  Fit(),
		Height: Fit(),
		Values: make(map[string]any),
	}
}

// ============================================================================
// Element — Node in Element Tree
// ============================================================================

// Element represents a node in the UI element tree.
//
// Field Usage by ElementType:
//
// ElementBox — Container:
//   - Uses: Type, Children, Layout, Style
//   - Renders: Background fill, border, arranged children
//   - Ignores: Text, Markdown, MarkdownText, Wrap, Render, ContentBuilder, OverlayX/Y
//
// ElementText — Single line:
//   - Uses: Type, Text, Style
//   - Renders: Text on single line
//   - Ignores: Children, Markdown, Wrap, Layout, Render, ContentBuilder, OverlayX/Y
//   - Note: Width determined by text length; height is 1
//
// ElementMultilineText — Multi-line text:
//   - Uses: Type, Text, Wrap, Style, Layout
//   - If Wrap=true: Text wraps at container width via reflow callback
//   - If Wrap=false: Text split on newlines; no word wrapping
//   - Renders: Multiple lines of text
//   - Ignores: Children, Markdown, MarkdownText, Render, ContentBuilder, OverlayX/Y
//
// ElementMarkdown — Rendered markdown:
//   - Uses: Type, MarkdownText, Markdown, Style, Layout
//   - MarkdownText: Raw markdown source (re-parsed on resize if non-empty)
//   - Markdown: Pre-parsed output (used if MarkdownText is empty)
//   - Renders: Multi-line text with per-character styling
//   - Ignores: Text, Wrap, Children, Render, ContentBuilder, OverlayX/Y
//
// ElementOverlay — Absolute positioned:
//   - Uses: Type, OverlayX, OverlayY, Children, Style
//   - Layout: Zero-size in flow; children painted at absolute position
//   - Renders: Children at (OverlayX, OverlayY) regardless of parent layout
//   - Ignores: Text, Markdown, Wrap, Layout, Render, ContentBuilder
//
// Hooks & Callbacks:
//   - Render: Hook function for component composition (called during render)
//   - ContentBuilder: Size-aware builder (called before layout with available dimensions)
//   - Props.Values: Custom state accessible to hooks
//
// Example (Box with children):
//
//	Element{
//	    Type: ElementBox,
//	    Layout: LayoutProps{Direction: Row, WidthSizing: Grow(1)},
//	    Children: []Element{...},
//	}
//
// Example (Text):
//
//	Element{
//	    Type: ElementText,
//	    Text: "Hello, world!",
//	    Style: Style{foreground: ColorGreen},
//	}
//
// Example (Markdown):
//
//	Element{
//	    Type: ElementMarkdown,
//	    MarkdownText: "# Title\n\n**Bold** text",
//	    Layout: LayoutProps{WidthSizing: Grow(1)},
//	}
type Element struct {
	// Id is an optional unique identifier for this element.
	// Not required; used for debugging or external tracking.
	Id string

	// Type identifies the element category (Box, Text, Markdown, etc.).
	// Determines which fields are active and how element behaves.
	Type ElementType

	// Key is an optional hint for element identity across renders.
	// Used by reconciliation to match elements (e.g., for lists).
	// Helps preserve element state (focus, scroll position) during updates.
	Key string

	// Text is the content for ElementText and ElementMultilineText.
	// Ignored by other element types.
	Text string

	// Wrap indicates whether ElementMultilineText should wrap at container width.
	// If true: text wraps with word boundaries, hard-breaks long words
	// If false: text split on newlines only, no word wrapping
	// Ignored by other element types.
	Wrap bool

	// Style carries colors, fonts, borders, background, etc.
	// Applied to this element and inherited by children (merged in render).
	// See Style type for available properties.
	Style Style

	// Layout contains sizing, positioning, and child arrangement properties.
	// Uses Direction (Row/Column), sizing modes (Fixed/Grow/Fit/Percent),
	// padding, gap, alignment, and justify.
	// Mirrors layout.go exactly; this is what the layout algorithm uses.
	Layout LayoutProps

	// Children is the list of child elements for ElementBox and ElementOverlay.
	// Ignored by text elements (ElementText, ElementMultilineText, ElementMarkdown).
	Children []Element

	// Render is an optional hook function for component composition.
	// If set, called during render pass to allow component logic.
	// Receives the Element itself as props.
	// Return value may be used for recursive rendering (implementation-dependent).
	Render func(props Element) Element

	// Props carries layout configuration (legacy) and custom hook state.
	// Prefer Element.Layout for layout properties; Props.Values for hook state.
	// See Props documentation for details.
	Props Props

	// Markdown is the pre-parsed markdown output (lines with styling).
	// Populated by render engine; read-only from outside.
	// Do NOT set directly; set MarkdownText instead.
	Markdown MarkdownContent

	// MarkdownText is the raw markdown source for ElementMarkdown.
	// If non-empty, render engine re-parses it and updates Markdown.
	// If empty, Markdown field is used as-is.
	// Allows content to update markdown dynamically (e.g., on resize).
	MarkdownText string

	// ContentBuilder is an optional size-aware content builder.
	// Called before layout with allocated width/height.
	// Returns an Element that replaces this node for rendering.
	// Allows responsive UI that changes structure based on available space.
	// If set, this element's children are ignored until ContentBuilder returns.
	ContentBuilder func(width, height int) Element

	// OverlayX is the absolute column (left edge) for ElementOverlay.
	// Only used when Type == ElementOverlay.
	// Screen-relative (not parent-relative).
	// Negative values allowed (off-screen overlay renders nothing).
	OverlayX int

	// OverlayY is the absolute row (top edge) for ElementOverlay.
	// Only used when Type == ElementOverlay.
	// Screen-relative (not parent-relative).
	// Negative values allowed (off-screen overlay renders nothing).
	OverlayY int
}

// ============================================================================
// LayoutProps — Layout Configuration
// ============================================================================

// LayoutProps contains the layout configuration for an element.
//
// LayoutProps mirrors layout.go exactly and is the authoritative source
// of layout properties used by the layout algorithm. Never modify the
// layout algorithm or layout.go without updating LayoutProps.
//
// Properties:
//   - Direction: Row (left-to-right) or Column (top-to-bottom)
//   - WidthSizing, HeightSizing: Sizing modes (Fixed, Grow, Fit, Percent)
//   - Padding: Interior spacing [top, right, bottom, left]
//   - Gap: Spacing between adjacent children
//   - Align: Cross-axis alignment (Start, Center, End, Stretch)
//   - Justify: Main-axis distribution (Start, End, Center, SpaceBetween, SpaceAround)
//
// Sizing Modes:
//   - SizingFixed(n): Exactly n pixels/columns
//   - SizingGrow(weight): Claim remaining space (weight determines share)
//   - SizingFit(): Shrink-wrap to children's intrinsic size
//   - SizingPercent(n): Percentage (0-100) of parent's resolved size
//
// Alignment (Cross-Axis):
//   - AlignStart: Start of cross axis (top in Row, left in Column)
//   - AlignCenter: Centered on cross axis
//   - AlignEnd: End of cross axis (bottom in Row, right in Column)
//   - AlignStretch: Full cross-axis size (default)
//
// Justify (Main-Axis Distribution):
//   - JustifyStart: Pack at start, extra space at end (default)
//   - JustifyEnd: Pack at end, extra space at start
//   - JustifyCenter: Center children, split extra space equally
//   - JustifySpaceBetween: Equal space between children
//   - JustifySpaceAround: Equal space around each child
//
// Example:
//
//	LayoutProps{
//	    Direction:    Row,
//	    WidthSizing:  Grow(1),
//	    HeightSizing: Fixed(3),
//	    PaddingTop:    1,
//	    PaddingRight:  2,
//	    PaddingBottom: 1,
//	    PaddingLeft:   2,
//	    Gap:           1,
//	    Align:         AlignCenter,
//	    Justify:       JustifySpaceBetween,
//	}
type LayoutProps struct {
	// Direction is the primary layout axis.
	Direction Direction
	// WidthSizing determines how width is computed.
	WidthSizing Sizing
	// HeightSizing determines how height is computed.
	HeightSizing Sizing
	// PaddingTop is interior spacing at top.
	PaddingTop int
	// PaddingRight is interior spacing at right.
	PaddingRight int
	// PaddingBottom is interior spacing at bottom.
	PaddingBottom int
	// PaddingLeft is interior spacing at left.
	PaddingLeft int

	// MarginTop is exterior spacing at top — space between this element and
	// whatever is above it, outside its border box.
	MarginTop int
	// MarginRight is exterior spacing at right.
	MarginRight int
	// MarginBottom is exterior spacing at bottom.
	MarginBottom int
	// MarginLeft is exterior spacing at left.
	MarginLeft int

	// Gap is minimum spacing between adjacent children on main axis.
	Gap int
	// Align controls child alignment on cross axis.
	Align Alignment
	// Justify controls space distribution along main axis.
	Justify Justify
}

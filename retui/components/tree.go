package components

import "github.com/subhasundardass/retui/retui"

// ─── Clean Tree Component ─────────────────────────────────────────────────────

type TreeNode struct {
	Label    string
	ID       string
	Children []TreeNode
}

type visibleNode struct {
	id     string
	label  string
	depth  int
	isLeaf bool
	isLast bool
}

func Tree(
	treeID string,
	nodes []TreeNode,
	focused bool,
	onChange func(id string),
) retui.Element {

	focusedIdx, setFocusedIdx := retui.UseState(0)
	expanded, setExpanded := retui.UseStateKeyed(treeID+":expanded", map[string]bool{})

	// Build visible tree
	visible := flattenTree(nodes, expanded)

	if len(visible) == 0 {
		return retui.Text("(empty)", retui.NewStyle())
	}

	if focusedIdx >= len(visible) {
		setFocusedIdx(len(visible) - 1)
		focusedIdx = len(visible) - 1
	}

	curr := visible[focusedIdx]

	// KEYBOARD NAVIGATION
	if focused {
		switch retui.CurrentKey.Code {
		case retui.KeyUp:
			if focusedIdx > 0 {
				setFocusedIdx(focusedIdx - 1)
			}
		case retui.KeyDown:
			if focusedIdx < len(visible)-1 {
				setFocusedIdx(focusedIdx + 1)
			}
		case retui.KeyEnter:
			if curr.isLeaf {
				if onChange != nil {
					onChange(curr.id)
				}
			} else {
				next := make(map[string]bool)
				for k, v := range expanded {
					next[k] = v
				}
				next[curr.id] = !expanded[curr.id]
				setExpanded(next)

				if onChange != nil {
					onChange(curr.id)
				}
			}
		}
	}

	// RENDER TREE LINES
	elems := make([]retui.Element, 0, len(visible))

	for i, node := range visible {
		isFocused := focused && i == focusedIdx

		// Build tree prefix (indentation with lines)
		prefix := ""
		if node.depth > 0 {
			for j := 0; j < node.depth-1; j++ {
				prefix += "│   "
			}
			if node.isLast {
				prefix += "└── "
			} else {
				prefix += "├── "
			}
		}

		// Build line with separate styles for prefix and content
		var lineElems []retui.Element

		// Prefix (tree lines) in light gray
		if prefix != "" {
			prefixStyle := retui.NewStyle().Foreground(retui.Cyan)
			lineElems = append(lineElems, retui.Text(prefix, prefixStyle))
		}

		// Icon + Label with focus styling
		contentStyle := retui.NewStyle()
		if isFocused {
			contentStyle = contentStyle.Background(retui.Blue).Foreground(retui.White).Bold(true)
		}
		// lineElems = append(lineElems, retui.Text(icon+node.label, contentStyle))
		lineElems = append(lineElems, retui.Text(node.label, contentStyle))

		// Combine as single row element
		elems = append(elems, retui.Box(
			retui.Props{Direction: retui.Row},
			retui.NewStyle(),
			lineElems...,
		))
	}

	return retui.Box(
		retui.Props{Direction: retui.Column},
		retui.NewStyle(),
		elems...,
	)
}

// Flatten tree into visible nodes based on expand state
func flattenTree(nodes []TreeNode, expanded map[string]bool) []visibleNode {
	var out []visibleNode
	walkTree(nodes, 0, expanded, &out)
	return out
}

func walkTree(nodes []TreeNode, depth int, expanded map[string]bool, out *[]visibleNode) {
	for i, n := range nodes {
		id := n.ID
		if id == "" {
			id = n.Label
		}

		*out = append(*out, visibleNode{
			id:     id,
			label:  n.Label,
			depth:  depth,
			isLeaf: len(n.Children) == 0,
			isLast: i == len(nodes)-1,
		})

		if expanded[id] && len(n.Children) > 0 {
			walkTree(n.Children, depth+1, expanded, out)
		}
	}
}

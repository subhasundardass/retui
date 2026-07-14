package example

import (
	"github.com/subhasundardass/retui/retui"
	"github.com/subhasundardass/retui/retui/components"
)

var sidebarTree = []components.TreeNode{
	{
		ID:    "example",
		Label: "Example",
		Children: []components.TreeNode{
			{
				ID:    "basic-inputs",
				Label: "Basic Inputs",
			},
			{
				ID:    "list",
				Label: "List",
			},
			{
				ID:    "windows",
				Label: "Windows",
			},
			{
				ID:    "other",
				Label: "Other",
			},
			{
				ID:    "state",
				Label: "State",
			},
		},
	},
}

func Sidebar() retui.Element {

	var findLeafNodes func([]components.TreeNode)
	isFocused := retui.IsFocused("sidebar")
	isLeafNode := make(map[string]bool)

	findLeafNodes = func(nodes []components.TreeNode) {
		for _, node := range nodes {
			// If it has no children, it's a leaf node (the end of the branch)
			isLeafNode[node.ID] = len(node.Children) == 0

			if len(node.Children) > 0 {
				findLeafNodes(node.Children)
			}
		}
	}
	findLeafNodes(sidebarTree)

	return retui.Box(
		retui.Props{Direction: retui.Column, Padding: [4]int{1, 0, 0, 1}, Width: retui.Fixed(30), Gap: 0},
		retui.NewStyle().Border(retui.Border{
			Top: true, Right: true, Bottom: true, Left: true,
			Chars: retui.BorderRounded, Color: retui.Blue,
			Title: "Navigation",
		}).Background(retui.Black),

		// Sidebar tree panel
		retui.Box(
			retui.Props{Direction: retui.Column, Gap: 0},
			retui.NewStyle(),

			components.Tree(
				"sidebar",
				sidebarTree,
				isFocused,
				func(id string) {

					// Check if the clicked ID is a leaf node
					if !isLeafNode[id] {
						return
					}

					if isFocused {
						retui.Debug("*Sidebar selected -> ", id)
						retui.PushScreen(id)
						retui.SetFocus("content")
					}
				},
			),
		),
	)
}

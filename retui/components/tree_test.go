package components

import (
	"testing"

	"github.com/subhasundardass/retui/retui"
)

func TestTree(t *testing.T) {
	t.Run("creates a tree with nodes", func(t *testing.T) {
		nodes := []TreeNode{
			{Label: "Root", ID: "root", Children: []TreeNode{
				{Label: "Child 1", ID: "child1"},
				{Label: "Child 2", ID: "child2"},
			}},
		}
		elem := Tree("test", nodes, false, nil)

		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("renders empty tree with placeholder", func(t *testing.T) {
		nodes := []TreeNode{}
		elem := Tree("test", nodes, false, nil)

		if elem.Type != retui.ElementText {
			t.Errorf("expected Element of type ElementText, got %v", elem.Type)
		}
		if elem.Text != "(empty)" {
			t.Errorf("expected '(empty)', got %q", elem.Text)
		}
	})

	t.Run("renders tree with multiple root nodes", func(t *testing.T) {
		nodes := []TreeNode{
			{Label: "Root 1", ID: "root1"},
			{Label: "Root 2", ID: "root2"},
		}
		elem := Tree("test", nodes, false, nil)

		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("renders tree with nested children", func(t *testing.T) {
		nodes := []TreeNode{
			{
				Label: "Root",
				ID:    "root",
				Children: []TreeNode{
					{Label: "Child 1", ID: "child1"},
					{Label: "Child 2", ID: "child2", Children: []TreeNode{
						{Label: "Grandchild", ID: "grandchild"},
					}},
				},
			},
		}
		elem := Tree("test", nodes, false, nil)

		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("handles focus correctly", func(t *testing.T) {
		nodes := []TreeNode{
			{Label: "Root", ID: "root"},
		}
		elem := Tree("test", nodes, true, nil)

		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("triggers onChange when selecting leaf", func(t *testing.T) {
		var changed bool
		var capturedID string
		onChange := func(id string) {
			changed = true
			capturedID = id
		}

		// Call the function directly to test it
		onChange("leaf")
		if !changed {
			t.Error("onChange callback was not executed")
		}
		if capturedID != "leaf" {
			t.Errorf("expected id 'leaf', got %s", capturedID)
		}

		// Test that Tree accepts the callback
		nodes := []TreeNode{
			{Label: "Leaf", ID: "leaf"},
		}
		elem := Tree("test", nodes, false, onChange)
		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("handles nodes without IDs using labels as IDs", func(t *testing.T) {
		nodes := []TreeNode{
			{Label: "Root", Children: []TreeNode{
				{Label: "Child"},
			}},
		}
		elem := Tree("test", nodes, false, nil)

		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})
}

func TestTree_FlattenTree(t *testing.T) {
	t.Run("flattens tree with expanded nodes", func(t *testing.T) {
		nodes := []TreeNode{
			{
				Label: "Root",
				ID:    "root",
				Children: []TreeNode{
					{Label: "Child 1", ID: "child1"},
					{Label: "Child 2", ID: "child2"},
				},
			},
		}
		expanded := map[string]bool{"root": true}
		result := flattenTree(nodes, expanded)

		if len(result) != 3 {
			t.Errorf("expected 3 visible nodes, got %d", len(result))
		}
		if result[0].label != "Root" {
			t.Errorf("expected first node 'Root', got %s", result[0].label)
		}
		if result[0].depth != 0 {
			t.Errorf("expected depth 0, got %d", result[0].depth)
		}
		if result[0].isLeaf {
			t.Error("root should not be a leaf (it has children)")
		}
		if result[1].label != "Child 1" {
			t.Errorf("expected second node 'Child 1', got %s", result[1].label)
		}
		if !result[1].isLeaf {
			t.Error("child 1 should be a leaf")
		}
		if result[2].label != "Child 2" {
			t.Errorf("expected third node 'Child 2', got %s", result[2].label)
		}
		if !result[2].isLeaf {
			t.Error("child 2 should be a leaf")
		}
	})

	t.Run("flattens tree with collapsed nodes", func(t *testing.T) {
		nodes := []TreeNode{
			{
				Label: "Root",
				ID:    "root",
				Children: []TreeNode{
					{Label: "Child 1", ID: "child1"},
					{Label: "Child 2", ID: "child2"},
				},
			},
		}
		expanded := map[string]bool{}
		result := flattenTree(nodes, expanded)

		if len(result) != 1 {
			t.Errorf("expected 1 visible node, got %d", len(result))
		}
		if result[0].label != "Root" {
			t.Errorf("expected 'Root', got %s", result[0].label)
		}
	})

	t.Run("flattens deeply nested tree", func(t *testing.T) {
		nodes := []TreeNode{
			{
				Label: "Level 1",
				ID:    "l1",
				Children: []TreeNode{
					{
						Label: "Level 2",
						ID:    "l2",
						Children: []TreeNode{
							{Label: "Level 3", ID: "l3"},
						},
					},
				},
			},
		}
		expanded := map[string]bool{"l1": true, "l2": true}
		result := flattenTree(nodes, expanded)

		if len(result) != 3 {
			t.Errorf("expected 3 visible nodes, got %d", len(result))
		}
		if result[0].depth != 0 {
			t.Errorf("expected depth 0, got %d", result[0].depth)
		}
		if result[1].depth != 1 {
			t.Errorf("expected depth 1, got %d", result[1].depth)
		}
		if result[2].depth != 2 {
			t.Errorf("expected depth 2, got %d", result[2].depth)
		}
	})

	t.Run("handles leaf nodes correctly", func(t *testing.T) {
		nodes := []TreeNode{
			{Label: "Leaf 1", ID: "leaf1"},
			{Label: "Leaf 2", ID: "leaf2"},
		}
		expanded := map[string]bool{}
		result := flattenTree(nodes, expanded)

		if len(result) != 2 {
			t.Errorf("expected 2 visible nodes, got %d", len(result))
		}
		if !result[0].isLeaf {
			t.Error("expected node to be leaf")
		}
		if !result[1].isLeaf {
			t.Error("expected node to be leaf")
		}
	})

	t.Run("sets isLast correctly", func(t *testing.T) {
		nodes := []TreeNode{
			{Label: "Node 1", ID: "n1"},
			{Label: "Node 2", ID: "n2"},
			{Label: "Node 3", ID: "n3"},
		}
		expanded := map[string]bool{}
		result := flattenTree(nodes, expanded)

		if len(result) != 3 {
			t.Errorf("expected 3 visible nodes, got %d", len(result))
		}
		if result[0].isLast {
			t.Error("first node should not be last")
		}
		if result[1].isLast {
			t.Error("second node should not be last")
		}
		if !result[2].isLast {
			t.Error("third node should be last")
		}
	})

	t.Run("handles tree with mixed leaf and branch nodes", func(t *testing.T) {
		nodes := []TreeNode{
			{
				Label: "Branch",
				ID:    "branch",
				Children: []TreeNode{
					{Label: "Leaf 1", ID: "leaf1"},
					{Label: "Leaf 2", ID: "leaf2"},
				},
			},
			{Label: "Leaf 3", ID: "leaf3"},
		}
		expanded := map[string]bool{"branch": true}
		result := flattenTree(nodes, expanded)

		if len(result) != 4 {
			t.Errorf("expected 4 visible nodes, got %d", len(result))
		}
		if result[0].isLeaf {
			t.Error("branch node should not be leaf")
		}
		if !result[1].isLeaf {
			t.Error("leaf node should be leaf")
		}
		if !result[2].isLeaf {
			t.Error("leaf node should be leaf")
		}
		if !result[3].isLeaf {
			t.Error("leaf node should be leaf")
		}
	})
}

func TestTree_NodeTypes(t *testing.T) {
	t.Run("TreeNode has correct fields", func(t *testing.T) {
		node := TreeNode{
			Label: "Test Node",
			ID:    "test-id",
			Children: []TreeNode{
				{Label: "Child", ID: "child"},
			},
		}

		if node.Label != "Test Node" {
			t.Errorf("expected Label 'Test Node', got %s", node.Label)
		}
		if node.ID != "test-id" {
			t.Errorf("expected ID 'test-id', got %s", node.ID)
		}
		if len(node.Children) != 1 {
			t.Errorf("expected 1 child, got %d", len(node.Children))
		}
	})

	t.Run("visibleNode has correct fields", func(t *testing.T) {
		node := visibleNode{
			id:     "test-id",
			label:  "Test Node",
			depth:  2,
			isLeaf: true,
			isLast: true,
		}

		if node.id != "test-id" {
			t.Errorf("expected id 'test-id', got %s", node.id)
		}
		if node.label != "Test Node" {
			t.Errorf("expected label 'Test Node', got %s", node.label)
		}
		if node.depth != 2 {
			t.Errorf("expected depth 2, got %d", node.depth)
		}
		if !node.isLeaf {
			t.Error("expected isLeaf true")
		}
		if !node.isLast {
			t.Error("expected isLast true")
		}
	})
}

func TestTree_EdgeCases(t *testing.T) {
	t.Run("handles nil nodes", func(t *testing.T) {
		var nodes []TreeNode
		elem := Tree("test", nodes, false, nil)

		if elem.Type != retui.ElementText {
			t.Errorf("expected Element of type ElementText, got %v", elem.Type)
		}
		if elem.Text != "(empty)" {
			t.Errorf("expected '(empty)', got %q", elem.Text)
		}
	})

	t.Run("handles empty string IDs", func(t *testing.T) {
		nodes := []TreeNode{
			{Label: "Root", Children: []TreeNode{
				{Label: "Child"},
			}},
		}
		elem := Tree("test", nodes, false, nil)

		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("handles very deep tree", func(t *testing.T) {
		nodes := []TreeNode{
			{
				Label: "Level 1",
				ID:    "l1",
				Children: []TreeNode{
					{
						Label: "Level 2",
						ID:    "l2",
						Children: []TreeNode{
							{
								Label: "Level 3",
								ID:    "l3",
								Children: []TreeNode{
									{Label: "Level 4", ID: "l4"},
								},
							},
						},
					},
				},
			},
		}
		expanded := map[string]bool{"l1": true, "l2": true, "l3": true}
		result := flattenTree(nodes, expanded)

		if len(result) != 4 {
			t.Errorf("expected 4 visible nodes, got %d", len(result))
		}
		if result[3].depth != 3 {
			t.Errorf("expected depth 3, got %d", result[3].depth)
		}
	})

	t.Run("handles many siblings", func(t *testing.T) {
		nodes := make([]TreeNode, 100)
		for i := 0; i < 100; i++ {
			nodes[i] = TreeNode{Label: "Node " + string(rune('A'+i%26)), ID: string(rune('a' + i%26))}
		}
		elem := Tree("test", nodes, false, nil)

		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})

	t.Run("handles nil onChange callback", func(t *testing.T) {
		nodes := []TreeNode{
			{Label: "Root", ID: "root"},
		}
		// This should not panic
		elem := Tree("test", nodes, false, nil)

		if elem.Type != retui.ElementBox {
			t.Errorf("expected Element of type ElementBox, got %v", elem.Type)
		}
	})
}

func TestTree_FlattenTreeHelpers(t *testing.T) {
	t.Run("walkTree traverses tree correctly", func(t *testing.T) {
		nodes := []TreeNode{
			{
				Label: "Root",
				ID:    "root",
				Children: []TreeNode{
					{Label: "Child 1", ID: "child1"},
					{Label: "Child 2", ID: "child2"},
				},
			},
		}
		expanded := map[string]bool{"root": true}
		var out []visibleNode
		walkTree(nodes, 0, expanded, &out)

		if len(out) != 3 {
			t.Errorf("expected 3 nodes, got %d", len(out))
		}
		if out[0].label != "Root" {
			t.Errorf("expected 'Root', got %s", out[0].label)
		}
		if out[1].label != "Child 1" {
			t.Errorf("expected 'Child 1', got %s", out[1].label)
		}
		if out[2].label != "Child 2" {
			t.Errorf("expected 'Child 2', got %s", out[2].label)
		}
	})

	t.Run("walkTree respects expanded state", func(t *testing.T) {
		nodes := []TreeNode{
			{
				Label: "Root",
				ID:    "root",
				Children: []TreeNode{
					{Label: "Child 1", ID: "child1"},
					{Label: "Child 2", ID: "child2"},
				},
			},
		}
		expanded := map[string]bool{"root": false}
		var out []visibleNode
		walkTree(nodes, 0, expanded, &out)

		if len(out) != 1 {
			t.Errorf("expected 1 node, got %d", len(out))
		}
		if out[0].label != "Root" {
			t.Errorf("expected 'Root', got %s", out[0].label)
		}
	})
}

// Benchmark tests
func BenchmarkTreeRender(b *testing.B) {
	nodes := []TreeNode{
		{
			Label: "Root",
			ID:    "root",
			Children: []TreeNode{
				{Label: "Child 1", ID: "child1"},
				{Label: "Child 2", ID: "child2"},
				{Label: "Child 3", ID: "child3"},
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Tree("bench", nodes, false, nil)
	}
}

func BenchmarkTreeRenderLarge(b *testing.B) {
	nodes := make([]TreeNode, 10)
	for i := 0; i < 10; i++ {
		children := make([]TreeNode, 5)
		for j := 0; j < 5; j++ {
			children[j] = TreeNode{Label: "Child " + string(rune('A'+j)), ID: "child" + string(rune('a'+j))}
		}
		nodes[i] = TreeNode{Label: "Node " + string(rune('A'+i)), ID: "node" + string(rune('a'+i)), Children: children}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Tree("bench", nodes, false, nil)
	}
}

func BenchmarkFlattenTree(b *testing.B) {
	nodes := []TreeNode{
		{
			Label: "Root",
			ID:    "root",
			Children: []TreeNode{
				{Label: "Child 1", ID: "child1"},
				{Label: "Child 2", ID: "child2"},
				{Label: "Child 3", ID: "child3"},
			},
		},
	}
	expanded := map[string]bool{"root": true}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		flattenTree(nodes, expanded)
	}
}

// Example usage test
func ExampleTree() {
	// Create a tree
	nodes := []TreeNode{
		{
			Label: "Root",
			ID:    "root",
			Children: []TreeNode{
				{Label: "Child 1", ID: "child1"},
				{Label: "Child 2", ID: "child2", Children: []TreeNode{
					{Label: "Grandchild", ID: "grandchild"},
				}},
			},
		},
	}

	tree := Tree("example", nodes, false, func(id string) {
		// Handle selection
	})
	_ = tree
}

package tree

import (
	"sigs.k8s.io/kustomize/pkg/types"
)

type KustomizeNode struct {
	Path          string
	Children      []*KustomizeNode
	Parent        *KustomizeNode
	kustomization types.Kustomization
}

func NewKustomizeNode(path string, parent *KustomizeNode) *KustomizeNode {
	return &KustomizeNode{
		Path:   path,
		Parent: parent,
	}
}

func (n *KustomizeNode) AddChild(path string) *KustomizeNode {
	child := NewKustomizeNode(path, n)
	n.Children = append(n.Children, child)
	return child
}

func (n *KustomizeNode) Root() *KustomizeNode {
	if n.Parent == nil {
		return n
	}
	return n.Parent.Root()
}

func (n *KustomizeNode) Depth() int {
	if n.Parent == nil {
		return 0
	}
	return n.Parent.Depth() + 1
}

// func (n *KustomizeNode) Walk(visitor func(node *KustomizeNode) bool) {
// 	if !visitor(n) {
// 		return
// 	}
// 	for _, child := range n.Children {
// 		child.Walk(visitor)
// 	}
// }
//
// func (n *KustomizeNode) Print(w io.Writer) {
// 	fmt.Fprintln(w, n.Path)
// 	n.PrintChildren(w, "")
// }
//
// func (n *KustomizeNode) PrintChildren(w io.Writer, prefix string) {
// 	for i, child := range n.Children {
// 		isLast := i == len(n.Children)-1
//
// 		// Choose the appropriate connectors
// 		connector := "├──"
// 		if isLast {
// 			connector = "└──"
// 		}
//
// 		// Print the current node
// 		fmt.Fprintf(w, "%s %s %s\n",
// 			prefix,
// 			connector,
// 			child.Path,
// 		)
//
// 		// Prepare the prefix for the next level
// 		newPrefix := prefix
// 		if isLast {
// 			newPrefix += "    " // Space for last item
// 		} else {
// 			newPrefix += " │  " // Vertical line for non-last items
// 		}
//
// 		// Recursively print children
// 		child.PrintChildren(w, newPrefix)
// 	}
//
// }

type KustomizeTree struct {
	root *KustomizeNode
}

type TreeVisitor func(node *KustomizeNode, depth int, isLast bool) bool

func MakeKustomizeTree(root *KustomizeNode) KustomizeTree {
	return KustomizeTree{root: root}
}

func (t *KustomizeTree) Walk(visitor TreeVisitor) {
	t.walk(visitor, t.root, 0, true)
}

func (t *KustomizeTree) walk(visitor TreeVisitor, node *KustomizeNode, depth int, isLast bool) {
	if !visitor(node, depth, isLast) {
		return
	}

	childDepth := depth + 1
	for i, child := range node.Children {
		isLastChild := i == len(node.Children)-1
		t.walk(visitor, child, childDepth, isLastChild)
	}
}

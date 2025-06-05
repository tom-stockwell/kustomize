package tree

import (
	"fmt"
	"io"
)

type PathNode struct {
	Path     string
	Children []*PathNode
	Parent   *PathNode
}

func NewPathNode(path string, parent *PathNode) *PathNode {
	return &PathNode{
		Path:   path,
		Parent: parent,
	}
}

func (n *PathNode) AddChild(path string) *PathNode {
	fmt.Println("adding:", path, "to", n.Path)
	child := NewPathNode(path, n)
	n.Children = append(n.Children, child)
	return child
}

func (n *PathNode) Root() *PathNode {
	if n.Parent == nil {
		return n
	}
	return n.Parent.Root()
}

func (n *PathNode) Depth() int {
	if n.Parent == nil {
		return 0
	}
	return n.Parent.Depth() + 1
}

func (n *PathNode) Walk(visitor func(node *PathNode) bool) {
	if !visitor(n) {
		return
	}
	for _, child := range n.Children {
		child.Walk(visitor)
	}
}

func (t *PathNode) Print(w io.Writer) {
	fmt.Fprintln(w, t.Path)
	t.PrintChildren(w, "")
}

func (t *PathNode) PrintChildren(w io.Writer, prefix string) {
	for i, child := range t.Children {
		isLast := i == len(t.Children)-1

		// Choose the appropriate connectors
		connector := "├──"
		if isLast {
			connector = "└──"
		}

		// Print the current node
		fmt.Fprintf(w, "%s %s %s\n",
			prefix,
			connector,
			child.Path,
		)

		// Prepare the prefix for the next level
		newPrefix := prefix
		if isLast {
			newPrefix += "    " // Space for last item
		} else {
			newPrefix += " │  " // Vertical line for non-last items
		}

		// Recursively print children
		child.PrintChildren(w, newPrefix)
	}

}

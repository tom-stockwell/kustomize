package tree

import (
	"fmt"
	"io"
)

type TreePrinter interface {
	PrintTree(tree KustomizeTree)
}

type SimpleTreePrinter struct {
	w io.Writer
}

func MakeSimpleTreePrinter(w io.Writer) TreePrinter {
	return SimpleTreePrinter{w: w}
}

func (p SimpleTreePrinter) PrintTree(tree KustomizeTree) {
	tree.Walk(p.print)
}

func (p SimpleTreePrinter) print(node *KustomizeTreeNode, depth int, isLast bool) bool {
	_, _ = fmt.Fprintln(p.w, node.Path)
	return true
}

type Style struct {
	Vertical   string
	Horizontal string
	Junction   string
	Corner     string
	Spacing    string
}

var BoxStyle = Style{
	Vertical:   "│",
	Horizontal: "──",
	Junction:   "├──",
	Corner:     "└──",
	Spacing:    " ",
}

type PrettyTreePrinter struct {
	w     io.Writer
	style Style
}

func MakePrettyTreePrinter(w io.Writer) TreePrinter {
	return PrettyTreePrinter{w: w, style: BoxStyle}
}

func (p PrettyTreePrinter) PrintTree(tree KustomizeTree) {
	p.print(tree.Root, 0, "")
}

func (p PrettyTreePrinter) print(n *KustomizeTreeNode, depth int, prefix string) {
	if depth == 0 {
		_, _ = fmt.Fprintln(p.w, n.Path)
	}

	childDepth := depth + 1
	for i, child := range n.Children {
		isLast := i == len(n.Children)-1

		connector := "├──"
		if isLast {
			connector = "└──"
		}

		_, _ = fmt.Fprintf(p.w, "%s %s %s\n", prefix, connector, child.Path)

		childPrefix := prefix
		if isLast {
			childPrefix += "    "
		} else {
			childPrefix += " │  "
		}

		// Recursively print children
		p.print(child, childDepth, childPrefix)
	}
}

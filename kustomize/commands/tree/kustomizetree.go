package tree

import (
	"sigs.k8s.io/kustomize/kyaml/filesys"
	"sigs.k8s.io/kustomize/pkg/types"
)

// KustomizeTreeNode

type KustomizeTreeNode struct {
	Path          string
	Children      []*KustomizeTreeNode
	Parent        *KustomizeTreeNode
	kustomization types.Kustomization
}

func NewKustomizeTreeNode(path string, parent *KustomizeTreeNode) *KustomizeTreeNode {
	return &KustomizeTreeNode{
		Path:   path,
		Parent: parent,
	}
}

func (n *KustomizeTreeNode) AddChild(child *KustomizeTreeNode) {
	child.Parent = n
	n.Children = append(n.Children, child)
}

func (n *KustomizeTreeNode) AddChildren(children []*KustomizeTreeNode) {
	for _, child := range children {
		n.AddChild(child)
	}
}

func (n *KustomizeTreeNode) Root() *KustomizeTreeNode {
	if n.Parent == nil {
		return n
	}
	return n.Parent.Root()
}

func (n *KustomizeTreeNode) Depth() int {
	if n.Parent == nil {
		return 0
	}
	return n.Parent.Depth() + 1
}

// KustomizeTree

type KustomizeTree struct {
	Root *KustomizeTreeNode
}

type TreeVisitor func(node *KustomizeTreeNode, depth int, isLast bool) bool

func NewKustomizeTree(path string, fSys filesys.FileSystem) (*KustomizeTree, error) {
	return newKustomizeTreeFactory(fSys).Build(path)
}

func (t *KustomizeTree) Walk(visitor TreeVisitor) {
	t.walk(visitor, t.Root, 0, true)
}

func (t *KustomizeTree) walk(visitor TreeVisitor, node *KustomizeTreeNode, depth int, isLast bool) {
	if !visitor(node, depth, isLast) {
		return
	}

	childDepth := depth + 1
	for i, child := range node.Children {
		isLastChild := i == len(node.Children)-1
		t.walk(visitor, child, childDepth, isLastChild)
	}
}

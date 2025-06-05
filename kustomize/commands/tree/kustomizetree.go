package tree

type TreeNode interface {
	Value() string
	Children() []TreeNode
	Parent() TreeNode
	Depth() int
	Root() TreeNode
}

type KustomizeTree[N TreeNode] struct {
	Root N
}

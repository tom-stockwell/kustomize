package tree

import (
	"net/url"
	"path/filepath"
	"strings"

	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/kyaml/filesys"
	"sigs.k8s.io/kustomize/pkg/git"
	"sigs.k8s.io/kustomize/pkg/types"
)

type KustomizeNode struct {
	Path          string
	Children      []*KustomizeNode
	Parent        *KustomizeNode
	kustomization types.Kustomization
}

type KustomizeTree struct {
	Root *KustomizeNode
}

type kustomizeTreeFactory struct {
	fSys filesys.FileSystem
	kp   *krusty.KustomizeParser
}

type TreeVisitor func(node *KustomizeNode, depth int, isLast bool) bool

func NewOrphanKustomizeNode(path string) *KustomizeNode {
	return &KustomizeNode{
		Path:   path,
		Parent: nil,
	}
}

func NewKustomizeNode(path string, parent *KustomizeNode) *KustomizeNode {
	return &KustomizeNode{
		Path:   path,
		Parent: parent,
	}
}

func (n *KustomizeNode) AddChild(child *KustomizeNode) *KustomizeNode {
	child.Parent = n
	n.Children = append(n.Children, child)
	return child
}

func (n *KustomizeNode) AddNewChild(path string) *KustomizeNode {
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

func NewKustomizeTree(path string, fSys filesys.FileSystem) (*KustomizeTree, error) {
	return newKustomizeTreeFactory(fSys).Build(path)
}

func newKustomizeTreeFactory(fSys filesys.FileSystem) *kustomizeTreeFactory {
	return &kustomizeTreeFactory{fSys, krusty.MakeKustomizerParser(krusty.MakeDefaultOptions())}
}

// KustomizeTree functions

func (t *KustomizeTree) Walk(visitor TreeVisitor) {
	t.walk(visitor, t.Root, 0, true)
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

// KustomizeTreeFactory functions

func (f *kustomizeTreeFactory) Build(path string) (*KustomizeTree, error) {
	root, err := f.parseKustomizationDir(path)
	if err != nil {
		return nil, err
	}
	return &KustomizeTree{root}, nil
}

func (f *kustomizeTreeFactory) parseKustomizationDir(path string) (*KustomizeNode, error) {
	k, err := f.kp.GetKustomization(f.fSys, path)
	if err != nil {
		return nil, err
	}

	node := NewOrphanKustomizeNode(path)

	// parse resources
	for _, r := range k.Resources {
		childNode, err := f.parseResource(r, path)
		if err != nil {
			return nil, err
		}
		node.AddChild(childNode)
	}

	// similarly parse bases
	for _, r := range k.Bases {
		childNode, err := f.parseResource(r, path)
		if err != nil {
			return nil, err
		}
		node.AddChild(childNode)
	}

	// similarly parse components
	for _, r := range k.Components {
		childNode, err := f.parseResource(r, path)
		if err != nil {
			return nil, err
		}
		node.AddChild(childNode)
	}

	// handle openapi ref
	if openApiPath, exists := k.OpenAPI["path"]; exists {
		node.AddNewChild(openApiPath)
	}

	// add references to patch files
	for _, p := range k.Patches {
		node.AddNewChild(p.Path)
	}
	for _, p := range k.PatchesStrategicMerge {
		node.AddNewChild(string(p))
	}
	for _, p := range k.PatchesJson6902 {
		node.AddNewChild(p.Path)
	}

	// handle configmap & secret generators
	for _, g := range k.SecretGenerator {
		if len(g.EnvSource) > 0 {
			node.AddNewChild(g.EnvSource)
		}
		for _, s := range g.EnvSources {
			node.AddNewChild(s)
		}
		for _, fs := range g.FileSources {
			kv := strings.SplitN(fs, "=", 2)
			if len(kv) == 2 {
				node.AddNewChild(kv[1])
			} else {
				node.AddNewChild(kv[0])
			}
		}
		node.AddNewChild(g.EnvSource)
	}

	for _, g := range k.ConfigMapGenerator {
		if len(g.EnvSource) > 0 {
			node.AddNewChild(g.EnvSource)
		}
		for _, s := range g.EnvSources {
			node.AddNewChild(s)
		}
		for _, fs := range g.FileSources {
			kv := strings.SplitN(fs, "=", 2)
			if len(kv) == 2 {
				node.AddNewChild(kv[1])
			} else {
				node.AddNewChild(kv[0])
			}
		}
	}

	return node, nil
}

func IsRemoteFile(path string) bool {
	u, err := url.Parse(path)
	return err == nil && (u.Scheme == "http" || u.Scheme == "https")
}

func (f *kustomizeTreeFactory) parseResource(path string, parentPath string) (*KustomizeNode, error) {
	_, err := git.NewRepoSpecFromUrl(path)
	if err == nil {
		return f.parseGitRepo(path)
	}
	if IsRemoteFile(path) {
		return f.parseRemoteFile(path)
	}
	return f.parseLocalFSys(path, parentPath)
}

func (f *kustomizeTreeFactory) parseRemoteFile(path string) (*KustomizeNode, error) {
	return NewOrphanKustomizeNode(path), nil
}

func (f *kustomizeTreeFactory) parseGitRepo(path string) (*KustomizeNode, error) {
	return f.parseRemoteFile(path)
}

func (f *kustomizeTreeFactory) parseLocalFSys(path string, parentPath string) (*KustomizeNode, error) {
	actualPath := path
	if !filepath.IsAbs(path) {
		actualPath = filepath.Join(parentPath, path)
	}
	if f.fSys.IsDir(actualPath) {
		return f.parseKustomizationDir(actualPath)
	}
	return NewOrphanKustomizeNode(actualPath), nil
}

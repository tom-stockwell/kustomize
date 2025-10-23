package krusty

import (
	"path"
	"path/filepath"

	"github.com/xlab/treeprint"
	"sigs.k8s.io/kustomize/api/internal/ls"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

type KustomizeFileTreeOptions struct {
	TreeStyle bool
}

type KustomizeFileTreeBuilder struct {
	ldr  *ls.KustomizationLoader
	opts *KustomizeFileTreeOptions
}

func NewKustomizeFileTreeBuilder(o *Options, opts *KustomizeFileTreeOptions) *KustomizeFileTreeBuilder {
	return &KustomizeFileTreeBuilder{
		ldr:  ls.MakeKustomizationLoader(o.LoadRestrictions, o.PluginConfig),
		opts: opts,
	}
}

func (fl *KustomizeFileTreeBuilder) Run(fSys filesys.FileSystem, path string) (treeprint.Tree, error) {
	rootTree := treeprint.New()
	kustTree, err := fl.buildKustTree(fSys, path, path, rootTree)
	if err != nil {
		return nil, err
	}
	if fl.opts.TreeStyle {
		return kustTree, nil
	} else {
		return rootTree, nil
	}
}

func (fl *KustomizeFileTreeBuilder) buildKustTree(fSys filesys.FileSystem, nodePath string, origRef string, parent treeprint.Tree) (treeprint.Tree, error) {
	k, err := fl.loadKustomization(fSys, nodePath)
	if err != nil {
		return nil, err
	}

	var kustFilePath string
	if fl.opts.TreeStyle {
		kustFilePath = filepath.Join(origRef, k.KustFileName)
	} else {
		kustFilePath = filepath.Join(k.Path, k.KustFileName)
	}

	tree := parent.AddBranch(kustFilePath)
	for _, f := range k.AllFiles() {
		if fl.opts.TreeStyle {
			tree.AddNode(f)
		} else {
			tree.AddNode(filepath.Join(nodePath, f))
		}
	}

	for _, d := range k.AllChildKustomizations() {
		_, err = fl.buildKustTree(fSys, path.Join(nodePath, d), d, tree)
		if err != nil {
			return nil, err
		}
	}
	return tree, nil
}

func (fl *KustomizeFileTreeBuilder) loadKustomization(fSys filesys.FileSystem, path string) (*ls.KustomizationWrapper, error) {
	k, kf, err := fl.ldr.LoadKustomization(fSys, path)
	if err != nil {
		return nil, err
	}
	return ls.MakeKustomizationWrapper(k, fSys, kf, path), nil
}

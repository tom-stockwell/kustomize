package tree

import (
	"net/url"
	"path/filepath"

	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/kyaml/filesys"
	"sigs.k8s.io/kustomize/pkg/git"
)

type kustomizeTreeFactory struct {
	fSys filesys.FileSystem
	kp   *krusty.KustomizeParser
}

func newKustomizeTreeFactory(fSys filesys.FileSystem) *kustomizeTreeFactory {
	return &kustomizeTreeFactory{fSys, krusty.MakeKustomizerParser(krusty.MakeDefaultOptions())}
}

func (f *kustomizeTreeFactory) Build(path string) (*KustomizeTree, error) {
	root, err := f.parseKustomizationDir(path)
	if err != nil {
		return nil, err
	}
	return &KustomizeTree{root}, nil
}

func (f *kustomizeTreeFactory) parseKustomizationDir(path string) (*KustomizeTreeNode, error) {
	k, err := f.kp.GetKustomization(f.fSys, path)
	if err != nil {
		return nil, err
	}
	kf := MakeKustomizationFiles(*k)

	node := NewKustomizeTreeNode(path, nil)

	// add files that potentially need to be recursively parsed as kustomizations
	allResources := make([]string, 0)
	allResources = append(allResources, kf.Resources...)
	allResources = append(allResources, kf.Components...)
	allResources = append(allResources, kf.Bases...)

	err = f.addChildNodes(node, allResources, path, true)
	if err != nil {
		return nil, err
	}

	// add all other file references that don't require recursion
	otherResources := make([]string, 0)
	if kf.OpenAPIPath != "" {
		otherResources = append(otherResources, kf.OpenAPIPath)
	}
	otherResources = append(otherResources, kf.Crds...)
	otherResources = append(otherResources, kf.Patches...)
	otherResources = append(otherResources, kf.PatchesStrategicMerge...)
	otherResources = append(otherResources, kf.PatchesJson6902...)
	otherResources = append(otherResources, kf.Configurations...)
	otherResources = append(otherResources, kf.Generators...)
	otherResources = append(otherResources, kf.Transformers...)
	otherResources = append(otherResources, kf.Validators...)
	otherResources = append(otherResources, kf.ConfigMapGeneratorFiles...)
	otherResources = append(otherResources, kf.SecretGeneratorFiles...)

	err = f.addChildNodes(node, otherResources, path, false)
	if err != nil {
		return nil, err
	}

	return node, nil
}

func (f *kustomizeTreeFactory) addChildNodes(parent *KustomizeTreeNode, paths []string, parentPath string, recurse bool) error {
	for _, path := range paths {
		err := f.addChildNode(parent, path, parentPath, recurse)
		if err != nil {
			return err
		}
	}
	return nil
}

func (f *kustomizeTreeFactory) addChildNode(parent *KustomizeTreeNode, path string, parentPath string, recurse bool) error {
	childNode, err := f.parseResource(path, parentPath, recurse)
	if err != nil {
		return err
	}
	parent.AddChild(childNode)
	return nil
}

func IsRemoteFile(path string) bool {
	u, err := url.Parse(path)
	return err == nil && (u.Scheme == "http" || u.Scheme == "https")
}

func (f *kustomizeTreeFactory) parseResource(path string, parentPath string, recurse bool) (*KustomizeTreeNode, error) {
	_, err := git.NewRepoSpecFromUrl(path)
	if err == nil {
		return f.parseGitRepo(path)
	}
	if IsRemoteFile(path) {
		return f.parseRemoteFile(path)
	}
	return f.parseLocalFSys(path, parentPath, recurse)
}

func (f *kustomizeTreeFactory) parseRemoteFile(path string) (*KustomizeTreeNode, error) {
	return NewKustomizeTreeNode(path, nil), nil
}

func (f *kustomizeTreeFactory) parseGitRepo(path string) (*KustomizeTreeNode, error) {
	return f.parseRemoteFile(path)
}

func (f *kustomizeTreeFactory) parseLocalFSys(path string, parentPath string, recurse bool) (*KustomizeTreeNode, error) {
	actualPath := path
	if !filepath.IsAbs(path) {
		actualPath = filepath.Join(parentPath, path)
	}
	if recurse && f.fSys.IsDir(actualPath) {
		return f.parseKustomizationDir(actualPath)
	}
	return NewKustomizeTreeNode(actualPath, nil), nil
}

// todo: guard against infinite recursion

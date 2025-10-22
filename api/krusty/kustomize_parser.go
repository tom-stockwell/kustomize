package krusty

import (
	"net/url"
	"path"
	"strings"

	"github.com/xlab/treeprint"
	fLdr "sigs.k8s.io/kustomize/api/internal/loader"
	pLdr "sigs.k8s.io/kustomize/api/internal/plugins/loader"
	"sigs.k8s.io/kustomize/api/internal/target"
	"sigs.k8s.io/kustomize/api/provider"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

type KustomizeParser struct {
	options        *Options
	depProvider    *provider.DepProvider
	loadRestrictor fLdr.LoadRestrictorFunc
	fSys           filesys.FileSystem
}

func MakeKustomizerParser(o *Options) *KustomizeParser {
	dp := provider.NewDepProvider()
	lr := fLdr.RestrictionNone
	if o.LoadRestrictions == types.LoadRestrictionsRootOnly {
		lr = fLdr.RestrictionRootOnly
	}
	fs := filesys.MakeFsOnDisk()
	return &KustomizeParser{
		depProvider:    dp,
		options:        o,
		loadRestrictor: lr,
		fSys:           fs,
	}
}

func (kp *KustomizeParser) LoadKustomization(path string) (*types.Kustomization, string, error) {
	resmapFactory := resmap.NewFactory(kp.depProvider.GetResourceFactory())

	ldr, err := fLdr.NewLoader(kp.loadRestrictor, path, kp.fSys)

	if err != nil {
		return nil, "", err
	}
	defer ldr.Cleanup()

	kt := target.NewKustTarget(
		ldr,
		kp.depProvider.GetFieldValidator(),
		resmapFactory,
		// The plugin configs are always located on disk, regardless of the fSys passed in
		pLdr.NewLoader(kp.options.PluginConfig, resmapFactory, filesys.MakeFsOnDisk()),
	)

	err = kt.Load()
	_, kustFileName, err := target.LoadKustFile(ldr)
	if err != nil {
		return nil, "", err
	}

	kust := kt.Kustomization()
	return &kust, kustFileName, nil
}

func (kp *KustomizeParser) kustFiles(k types.Kustomization, root string) []string {
	files := make([]string, 0)

	files = append(files, kp.filterResources(k.Resources, root, false)...)
	files = append(files, kp.filterResources(k.Bases, root, false)...)
	files = append(files, filesFromKvSource(getConfigMapKvSources(k.ConfigMapGenerator))...)
	files = append(files, filesFromKvSource(getSecretGenKvSources(k.SecretGenerator))...)
	files = append(files, kp.filterResources(k.Configurations, root, false)...)

	files = append(files, kp.filterResources(k.Generators, root, false)...)
	files = append(files, kp.filterResources(k.Transformers, root, false)...)
	files = append(files, kp.filterResources(k.Validators, root, false)...)

	for _, patch := range k.Patches {
		if patch.Path != "" {
			files = append(files, patch.Path)
		}
	}

	for _, patch := range k.PatchesStrategicMerge {
		files = append(files, string(patch))
	}

	for _, patch := range k.PatchesJson6902 {
		if patch.Path != "" {
			files = append(files, patch.Path)
		}
	}

	if openAPIPath, exists := k.OpenAPI["path"]; exists {
		files = append(files, openAPIPath)
	}

	return files
}

func getConfigMapKvSources(args []types.ConfigMapArgs) []types.KvPairSources {
	kvPairs := make([]types.KvPairSources, len(args))
	for _, cm := range args {
		kvPairs = append(kvPairs, cm.KvPairSources)
	}
	return kvPairs
}

func getSecretGenKvSources(args []types.SecretArgs) []types.KvPairSources {
	kvPairs := make([]types.KvPairSources, len(args))
	for _, sec := range args {
		kvPairs = append(kvPairs, sec.KvPairSources)
	}
	return kvPairs
}

func filesFromKvSource(sources []types.KvPairSources) []string {
	files := make([]string, 0)
	for _, s := range sources {
		files = append(files, s.EnvSources...)
		if s.EnvSource != "" {
			files = append(files, s.EnvSource)
		}

		for _, fileSource := range s.FileSources {
			// parse "key=file" or just "file"
			parts := strings.SplitN(fileSource, "=", 2)
			files = append(files, parts[len(parts)-1])
		}
	}
	return files
}

func (kp *KustomizeParser) kustDirs(k types.Kustomization, root string) []string {
	files := make([]string, 0)
	files = append(files, kp.filterResources(k.Resources, root, true)...)
	files = append(files, kp.filterResources(k.Bases, root, true)...)
	files = append(files, kp.filterResources(k.Components, root, true)...)
	return files
}

func IsRemoteFile(path string) bool {
	u, err := url.Parse(path)
	return err == nil && (u.Scheme == "http" || u.Scheme == "https")
}

func (kp *KustomizeParser) filterResources(raw []string, root string, isDir bool) []string {
	matching := make([]string, 0)
	for _, f := range raw {
		relToRoot := path.Join(root, f)
		if !IsRemoteFile(f) && kp.fSys.IsDir(relToRoot) == isDir {
			matching = append(matching, f)
		}
	}
	return matching
}

func (kp *KustomizeParser) BuildKustTreeFromRoot(path string) (*treeprint.Tree, error) {
	return kp.buildKustTree(path, nil)
}

func (kp *KustomizeParser) buildKustTree(nodePath string, parent treeprint.Tree) (*treeprint.Tree, error) {
	k, kf, err := kp.LoadKustomization(nodePath)

	if err != nil || k == nil {
		return nil, err
	}

	kustFilePath := path.Join(nodePath, kf)

	var tree treeprint.Tree
	if parent == nil {
		tree = treeprint.NewWithRoot(nodePath)
	} else {
		tree = parent.AddBranch(nodePath)
	}
	tree.AddNode(kustFilePath)

	for _, f := range kp.kustFiles(*k, nodePath) {
		tree.AddNode(path.Join(nodePath, f))
	}

	for _, d := range kp.kustDirs(*k, nodePath) {
		_, err = kp.buildKustTree(path.Join(nodePath, d), tree)
		if err != nil {
			return nil, err
		}
	}

	return &tree, nil

}

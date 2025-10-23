package ls

import (
	"fmt"
	"path/filepath"

	fLdr "sigs.k8s.io/kustomize/api/internal/loader"
	pLdr "sigs.k8s.io/kustomize/api/internal/plugins/loader"
	"sigs.k8s.io/kustomize/api/internal/target"
	"sigs.k8s.io/kustomize/api/konfig"
	"sigs.k8s.io/kustomize/api/provider"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

type KustomizationLoader struct {
	depProvider      *provider.DepProvider
	loadRestrictions types.LoadRestrictions
	pluginConfig     *types.PluginConfig
}

func MakeKustomizationLoader(loadRestrictions types.LoadRestrictions, pluginConfig *types.PluginConfig) *KustomizationLoader {
	return &KustomizationLoader{
		loadRestrictions: loadRestrictions,
		pluginConfig:     pluginConfig,
		depProvider:      provider.NewDepProvider(),
	}
}

func (l *KustomizationLoader) LoadKustomization(fSys filesys.FileSystem, path string) (types.Kustomization, string, error) {
	lr := fLdr.RestrictionNone
	if l.loadRestrictions == types.LoadRestrictionsRootOnly {
		lr = fLdr.RestrictionRootOnly
	}
	ldr, err := fLdr.NewLoader(lr, path, fSys)
	if err != nil {
		return types.Kustomization{}, "", err
	}

	kt := target.NewKustTarget(
		ldr,
		nil,
		nil,
		// The plugin configs are always located on disk, regardless of the fSys passed in
		// pLdr.NewLoader(o.PluginConfig, resmapFactory, filesys.MakeFsOnDisk()),
		pLdr.NewLoader(l.pluginConfig, nil, filesys.MakeFsOnDisk()),
	)
	defer ldr.Cleanup()

	err = kt.Load()
	if err != nil {
		return types.Kustomization{}, "", err
	}

	kustFileName, err := getKustFileName(fSys, path)
	if err != nil {
		return types.Kustomization{}, "", err
	}

	return kt.Kustomization(), kustFileName, nil
}

func getKustFileName(fSys filesys.FileSystem, root string) (string, error) {
	match := 0
	var kustFileName string
	for _, kf := range konfig.RecognizedKustomizationFileNames() {
		kustPath := filepath.Join(root, kf)
		if fSys.Exists(kustPath) && !fSys.IsDir(kustPath) {
			match += 1
			kustFileName = kf
		}
	}
	switch match {
	case 0:
		return "", target.NewErrMissingKustomization(root)
	case 1:
		return kustFileName, nil
	default:
		return "", fmt.Errorf("Found multiple kustomization files under: %s\n", root)
	}
}

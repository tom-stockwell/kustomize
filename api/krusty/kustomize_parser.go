package krusty

import (
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
	kustomizer     *Kustomizer
	loadRestrictor fLdr.LoadRestrictorFunc
}

func MakeKustomizerParser(o *Options) *KustomizeParser {
	dp := provider.NewDepProvider()
	lr := fLdr.RestrictionNone
	if o.LoadRestrictions == types.LoadRestrictionsRootOnly {
		lr = fLdr.RestrictionRootOnly
	}
	return &KustomizeParser{
		kustomizer:     MakeKustomizer(o),
		depProvider:    dp,
		options:        o,
		loadRestrictor: lr,
	}

}

func (kp *KustomizeParser) GetKustomization(fSys filesys.FileSystem, path string) (*types.Kustomization, error) {
	println("In kustomizer:GetKustomization")

	resmapFactory := resmap.NewFactory(kp.depProvider.GetResourceFactory())

	ldr, err := fLdr.NewLoader(kp.loadRestrictor, path, fSys)
	if err != nil {
		return nil, err
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
	if err != nil {
		return nil, err
	}

	kust := kt.Kustomization()
	return &kust, nil
}

package ls

import (
	"encoding/json"
	"path"
	"strings"

	"sigs.k8s.io/kustomize/api/internal/loader"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

type KustomizationWrapper struct {
	kustomization types.Kustomization
	fSys          filesys.FileSystem
	KustFileName  string
	Path          string
}

func MakeKustomizationWrapper(k types.Kustomization, fSys filesys.FileSystem, kf string, path string) *KustomizationWrapper {
	return &KustomizationWrapper{
		kustomization: k,
		fSys:          fSys,
		KustFileName:  kf,
		Path:          path,
	}
}

// Kustomization returns a copy of the immutable, internal kustomization object.
func (w KustomizationWrapper) Kustomization() types.Kustomization {
	var result types.Kustomization
	b, _ := json.Marshal(w.kustomization)
	json.Unmarshal(b, &result)
	return result
}

func (w KustomizationWrapper) AllFiles() []string {
	files := make([]string, 0)

	files = append(files, w.filterResources(w.kustomization.Resources, false)...)
	files = append(files, w.filterResources(w.kustomization.Bases, false)...)
	files = append(files, filesFromKvSource(getConfigMapKvSources(w.kustomization.ConfigMapGenerator))...)
	files = append(files, filesFromKvSource(getSecretGenKvSources(w.kustomization.SecretGenerator))...)
	files = append(files, w.filterResources(w.kustomization.Configurations, false)...)

	files = append(files, w.filterResources(w.kustomization.Generators, false)...)
	files = append(files, w.filterResources(w.kustomization.Transformers, false)...)
	files = append(files, w.filterResources(w.kustomization.Validators, false)...)

	for _, patch := range w.kustomization.Patches {
		if patch.Path != "" {
			files = append(files, patch.Path)
		}
	}

	for _, patch := range w.kustomization.PatchesStrategicMerge {
		files = append(files, string(patch))
	}

	for _, patch := range w.kustomization.PatchesJson6902 {
		if patch.Path != "" {
			files = append(files, patch.Path)
		}
	}

	if openAPIPath, exists := w.kustomization.OpenAPI["path"]; exists {
		files = append(files, openAPIPath)
	}

	return files
}

func (w KustomizationWrapper) AllChildKustomizations() []string {
	files := make([]string, 0)
	files = append(files, w.filterResources(w.kustomization.Resources, true)...)
	files = append(files, w.filterResources(w.kustomization.Bases, true)...)
	files = append(files, w.filterResources(w.kustomization.Components, true)...)
	return files
}

func (w *KustomizationWrapper) filterResources(raw []string, isDir bool) []string {
	matching := make([]string, 0)
	for _, f := range raw {
		if !loader.IsRemoteFile(f) && w.fSys.IsDir(path.Join(w.Path, f)) == isDir {
			matching = append(matching, f)
		}
	}
	return matching
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

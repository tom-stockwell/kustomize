package tree

import (
	"strings"

	"sigs.k8s.io/kustomize/api/types"
)

type KustomizationFiles struct {
	Resources               []string
	Components              []string
	Bases                   []string
	Crds                    []string
	OpenAPIPath             string
	Patches                 []string
	PatchesStrategicMerge   []string
	PatchesJson6902         []string
	Configurations          []string
	Generators              []string
	Transformers            []string
	Validators              []string
	ConfigMapGeneratorFiles []string
	SecretGeneratorFiles    []string
}

func MakeKustomizationFiles(k types.Kustomization) KustomizationFiles {
	kfr := KustomizationFiles{
		Resources:             make([]string, len(k.Resources)),
		Components:            make([]string, len(k.Components)),
		Bases:                 make([]string, len(k.Bases)),
		Crds:                  make([]string, len(k.Crds)),
		Configurations:        make([]string, len(k.Configurations)),
		Generators:            make([]string, len(k.Generators)),
		Transformers:          make([]string, len(k.Transformers)),
		Validators:            make([]string, len(k.Validators)),
		Patches:               make([]string, 0, len(k.Patches)),
		PatchesJson6902:       make([]string, 0, len(k.PatchesJson6902)),
		PatchesStrategicMerge: make([]string, 0, len(k.PatchesStrategicMerge)),
	}

	copy(kfr.Resources, k.Resources)
	copy(kfr.Components, k.Components)
	copy(kfr.Bases, k.Bases)
	copy(kfr.Crds, k.Crds)
	copy(kfr.Configurations, k.Configurations)
	copy(kfr.Generators, k.Generators)
	copy(kfr.Transformers, k.Transformers)
	copy(kfr.Validators, k.Validators)

	if openAPIPath, exists := k.OpenAPI["path"]; exists {
		kfr.OpenAPIPath = openAPIPath
	}

	for _, patch := range k.Patches {
		if patch.Path != "" {
			kfr.Patches = append(kfr.Patches, patch.Path)
		}
	}

	for _, patch := range k.PatchesStrategicMerge {
		kfr.PatchesStrategicMerge = append(kfr.PatchesStrategicMerge, string(patch))
	}

	for _, patch := range k.PatchesJson6902 {
		if patch.Path != "" {
			kfr.Patches = append(kfr.Patches, patch.Path)
		}
	}

	kfr.SecretGeneratorFiles = kfr.secretGeneratorFiles(k)
	kfr.ConfigMapGeneratorFiles = kfr.configMapGeneratorFiles(k)

	return kfr
}

func (kfr *KustomizationFiles) extractPatchFiles(k types.Kustomization) []string {
	var patchFiles []string

	// Extract from Patches
	for _, patch := range k.Patches {
		if patch.Path != "" {
			patchFiles = append(patchFiles, patch.Path)
		}
	}

	// Extract from deprecated PatchesStrategicMerge
	for _, patch := range k.PatchesStrategicMerge {
		patchFiles = append(patchFiles, string(patch))
	}

	// Extract from deprecated PatchesJson6902
	for _, patch := range k.PatchesJson6902 {
		if patch.Path != "" {
			patchFiles = append(patchFiles, patch.Path)
		}
	}

	return patchFiles
}

func (kfr *KustomizationFiles) configMapGeneratorFiles(k types.Kustomization) []string {
	files := make([]string, 0, len(k.ConfigMapGenerator))

	// ConfigMap generator files
	for _, cm := range k.ConfigMapGenerator {
		if cm.EnvSource != "" {
			files = append(files, cm.EnvSource)
		}

		files = append(files, cm.EnvSources...)

		for _, fileSource := range cm.FileSources {
			// parse "key=file" or just "file"
			parts := strings.SplitN(fileSource, "=", 2)
			if len(parts) == 2 {
				files = append(files, parts[1])
			} else {
				files = append(files, parts[0])
			}
		}
	}
	return files
}

func (kfr *KustomizationFiles) secretGeneratorFiles(k types.Kustomization) []string {
	files := make([]string, 0, len(k.SecretGenerator))

	// ConfigMap generator files
	for _, s := range k.SecretGenerator {
		if s.EnvSource != "" {
			files = append(files, s.EnvSource)
		}

		files = append(files, s.EnvSources...)

		for _, fileSource := range s.FileSources {
			// parse "key=file" or just "file"
			parts := strings.SplitN(fileSource, "=", 2)
			if len(parts) == 2 {
				files = append(files, parts[1])
			} else {
				files = append(files, parts[0])
			}
		}
	}
	return files
}

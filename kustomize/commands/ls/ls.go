// Copyright 2024 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

// Package ls implements the `kustomize ls` command.
package ls

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/api/konfig"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/filesys"
	"sigs.k8s.io/yaml"
)

type options struct {
	fSys   filesys.FileSystem
	writer io.Writer
}

// NewCmdLs returns a cobra command that lists all local files referenced by
// the kustomization at the given directory.
func NewCmdLs(fSys filesys.FileSystem, writer io.Writer) *cobra.Command {
	o := &options{fSys: fSys, writer: writer}

	return &cobra.Command{
		Use:   "ls [dir]",
		Short: "List files referenced by a kustomization",
		Long: `Recursively prints every local file referenced by the kustomization
in the given directory (default: current directory).

Paths are printed relative to the given directory.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				return o.run(args[0])
			}
			dir, err := os.Getwd()
			if err != nil {
				return err
			}
			return o.run(dir)
		},
	}
}

func (o *options) run(dir string) error {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return err
	}

	files, err := collectFiles(o.fSys, absDir, make(map[string]struct{}))
	if err != nil {
		return err
	}

	for _, f := range files {
		rel, err := filepath.Rel(absDir, f)
		if err != nil {
			fmt.Fprintln(o.writer, f)
		} else {
			fmt.Fprintln(o.writer, rel)
		}
	}
	return nil
}

func findKustomizationFile(fSys filesys.FileSystem, dir string) (string, error) {
	for _, name := range konfig.RecognizedKustomizationFileNames() {
		p := filepath.Join(dir, name)
		if fSys.Exists(p) && !fSys.IsDir(p) {
			return p, nil
		}
	}
	return "", fmt.Errorf("no kustomization file found in %s", dir)
}

// isRemoteRef returns true for URLs and remote repository references that
// kustomize supports (e.g. github.com/..., https://...).
func isRemoteRef(ref string) bool {
	return strings.Contains(ref, "://") ||
		strings.HasPrefix(ref, "github.com/") ||
		strings.HasPrefix(ref, "gitlab.com/") ||
		strings.HasPrefix(ref, "bitbucket.org/")
}

// collectFiles recursively collects all local files referenced by the
// kustomization rooted at dir. visited prevents infinite cycles.
func collectFiles(fSys filesys.FileSystem, dir string, visited map[string]struct{}) ([]string, error) {
	if _, ok := visited[dir]; ok {
		return nil, fmt.Errorf("cyclic reference detected: %s", dir)
	}
	visited[dir] = struct{}{}

	kFile, err := findKustomizationFile(fSys, dir)
	if err != nil {
		return nil, err
	}

	content, err := fSys.ReadFile(kFile)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", kFile, err)
	}

	var k types.Kustomization
	if err := yaml.Unmarshal(content, &k); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", kFile, err)
	}

	var files []string
	seen := make(map[string]struct{})

	addFile := func(path string) {
		if _, ok := seen[path]; !ok {
			seen[path] = struct{}{}
			files = append(files, path)
		}
	}

	processRef := func(ref string) error {
		if isRemoteRef(ref) {
			return nil
		}
		p := filepath.Join(dir, ref)
		if !fSys.Exists(p) {
			return nil
		}
		if fSys.IsDir(p) {
			sub, err := collectFiles(fSys, p, visited)
			if err != nil {
				return err
			}
			for _, f := range sub {
				addFile(f)
			}
		} else {
			addFile(p)
		}
		return nil
	}

	processFileSource := func(src string) error {
		if i := strings.Index(src, "="); i >= 0 {
			src = src[i+1:]
		}
		return processRef(src)
	}

	addFile(kFile)

	for _, r := range k.Resources {
		if err := processRef(r); err != nil {
			return nil, err
		}
	}

	for _, c := range k.Components {
		if err := processRef(c); err != nil {
			return nil, err
		}
	}

	for _, p := range k.Patches {
		if p.Path != "" {
			if err := processRef(p.Path); err != nil {
				return nil, err
			}
		}
	}

	for _, p := range k.PatchesStrategicMerge {
		s := string(p)
		// Inline patches contain newlines; file references do not.
		if !strings.Contains(s, "\n") {
			if err := processRef(s); err != nil {
				return nil, err
			}
		}
	}

	for _, p := range k.PatchesJson6902 {
		if p.Path != "" {
			if err := processRef(p.Path); err != nil {
				return nil, err
			}
		}
	}

	for _, list := range [][]string{
		k.Configurations,
		k.Generators,
		k.Transformers,
		k.Validators,
		k.Crds,
	} {
		for _, ref := range list {
			if err := processRef(ref); err != nil {
				return nil, err
			}
		}
	}

	for _, gen := range k.ConfigMapGenerator {
		for _, f := range gen.FileSources {
			if err := processFileSource(f); err != nil {
				return nil, err
			}
		}
		for _, f := range gen.EnvSources {
			if err := processRef(f); err != nil {
				return nil, err
			}
		}
	}

	for _, gen := range k.SecretGenerator {
		for _, f := range gen.FileSources {
			if err := processFileSource(f); err != nil {
				return nil, err
			}
		}
		for _, f := range gen.EnvSources {
			if err := processRef(f); err != nil {
				return nil, err
			}
		}
	}

	return files, nil
}

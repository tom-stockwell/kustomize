package ls

import (
	"fmt"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/api/konfig"
	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

var theArgs struct {
	kustomizationPath string
}

// options:
//   - full paths - (relative to cwd), otherwise relative to the parent
//   - show / hide / recurse remote files
//   - flat vs tree view
//    - kustomize cfg files --tree --full-paths --relative-paths --absolute-paths --show-remote --parse-remote

var theFlags struct {
	style OutputStyle
	// hideRemoteFiles bool
	// absolute paths
	// include kustomization directories, or just list kust files?
	// include type of resource (resource, component, configmap, etc. the file is)
}

func newFiletreeOptions() *krusty.KustomizeFileTreeOptions {
	return &krusty.KustomizeFileTreeOptions{
		TreeStyle: theFlags.style == OutputStyleTree,
	}
}

func Validate(args []string) error {
	if len(args) > 1 {
		return fmt.Errorf("specify one path to %s", konfig.DefaultKustomizationFileName())
	}
	theArgs.kustomizationPath = args[0]
	return nil
}

func NewLsCommand(fSys filesys.FileSystem) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "ls",
		Example: "kustomize ls",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := Validate(args); err != nil {
				return err
			}

			b := krusty.NewKustomizeFileTreeBuilder(krusty.MakeDefaultOptions(), newFiletreeOptions())
			filetree, err := b.Run(fSys, theArgs.kustomizationPath)
			if err != nil {
				return err
			}

			switch theFlags.style {
			case OutputStyleTree:
				fmt.Println(filetree.String())
			case OutputStyleList:
				filetree.VisitAll(flatTreePrinter)
			}

			return nil
		},
	}

	AddFlagStyle(cmd.Flags())

	return cmd
}

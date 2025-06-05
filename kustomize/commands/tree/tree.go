package tree

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/api/konfig"
	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/kyaml/filesys"
)

var theArgs struct {
	kustomizationPath string
}

func NewCmdTree(fSys filesys.FileSystem, w io.Writer) *cobra.Command {
	versionCmd := cobra.Command{
		Use:     "tree",
		Example: `kustomize tree`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := Validate(args); err != nil {
				return err
			}
			return RunTree(fSys, w, theArgs.kustomizationPath)
		},
	}

	return &versionCmd
}

func RunTree(fSys filesys.FileSystem, w io.Writer, path string) error {
	_, err := fmt.Fprintln(w, "test output")
	if err != nil {
		return err
	}

	kp := krusty.MakeKustomizerParser(krusty.MakeDefaultOptions())

	root := NewPathNode(path, nil)

	parseDir(fSys, kp, root)

	fmt.Println("Walking tree")
	root.Walk(func(node *PathNode) bool {
		fmt.Println(node.Path, len(node.Children))
		return true
	})
	fmt.Println(root)
	fmt.Println("TREE")
	root.Print(w)

	return nil
}

func parseDir(fSys filesys.FileSystem, kp *krusty.KustomizeParser, node *PathNode) error {
	k, err := kp.GetKustomization(fSys, node.Path)
	if err != nil {
		return err
	}

	fmt.Println("running for:", node.Path)

	for _, r := range k.Resources {
		rPath := filepath.Join(node.Path, r)
		newNode := node.AddChild(rPath)
		if fSys.IsDir(rPath) {
			err = parseDir(fSys, kp, newNode)
			if err != nil {
				fmt.Println("Error reading", rPath)
			}
		} else {
			fmt.Println(rPath, "is not dir")
		}
		fmt.Println(node)
	}

	return nil
}

// func parseFiles(string kustPath, []string paths) error {
//
// }

// Validate validates build command args and flags.
func Validate(args []string) error {
	if len(args) > 1 {
		return fmt.Errorf("specify one path to " + konfig.DefaultKustomizationFileName())
	}
	if len(args) == 0 {
		theArgs.kustomizationPath = filesys.SelfDir
	} else {
		theArgs.kustomizationPath = args[0]
	}
	return nil
}

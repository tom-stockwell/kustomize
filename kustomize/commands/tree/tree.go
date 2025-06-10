package tree

import (
	"fmt"
	"io"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/api/konfig"
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

	ktree, err := NewKustomizeTree(path, fSys)
	if err != nil {
		return err
	}

	fmt.Println("\nPrinting simple tree:")
	printer := MakeSimpleTreePrinter(w)
	printer.PrintTree(*ktree)

	fmt.Println("\nPrinting pretty tree:")
	printer = MakePrettyTreePrinter(w)
	printer.PrintTree(*ktree)

	return nil
}

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

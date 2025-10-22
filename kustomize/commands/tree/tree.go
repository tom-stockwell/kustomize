package tree

import (
	"fmt"
	"io"
	"path"

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

func getKustFile(fSys filesys.FileSystem, root string) (string, error) {
	match := 0
	var kustFileName string
	for _, kf := range konfig.RecognizedKustomizationFileNames() {
		fp := path.Join(root, kf)
		if fSys.Exists(fp) && !fSys.IsDir(fp) {
			match += 1
			kustFileName = kf
		}
	}
	switch match {
	case 0:
		return "", fmt.Errorf("No kustomization found under: %s\n", root)
	case 1:
		return kustFileName, nil
	default:
		return "", fmt.Errorf("Found multiple kustomization files under: %s\n", root)
	}
}

func RunTree(fSys filesys.FileSystem, w io.Writer, path string) error {
	thepath := "examples/springboot/overlays/production"
	kp := krusty.MakeKustomizerParser(krusty.MakeDefaultOptions())
	tree, err := kp.BuildKustTreeFromRoot(thepath)
	if err != nil {
		return err
	}

	// options:
	//   - full paths - (relative to cwd), otherwise relative to the parent
	//   - show / hide / recurse remote files
	//   - flat vs tree view
	//    - kustomize cfg files --tree --full-paths --relative-paths --absolute-paths --show-remote --parse-remote

	fmt.Println((*tree).String())
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

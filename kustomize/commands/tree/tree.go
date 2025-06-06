package tree

import (
	"fmt"
	"io"
	"net/url"
	"path/filepath"

	"github.com/spf13/cobra"
	"sigs.k8s.io/kustomize/api/konfig"
	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/kyaml/filesys"
	"sigs.k8s.io/kustomize/pkg/git"
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
	kp := krusty.MakeKustomizerParser(krusty.MakeDefaultOptions())

	root := NewKustomizeNode(path, nil)
	parseDir(fSys, kp, root)

	tree := MakeKustomizeTree(root)

	fmt.Println("\nPrinting simple tree:\n")
	printer := MakeSimpleTreePrinter(w)
	printer.PrintTree(tree)

	fmt.Println("\nPrinting pretty tree:\n")
	printer = MakePrettyTreePrinter(w)
	printer.PrintTree(tree)

	return nil
}

func IsRemoteFile(path string) bool {
	u, err := url.Parse(path)
	return err == nil && (u.Scheme == "http" || u.Scheme == "https")
}

func parseDir(fSys filesys.FileSystem, kp *krusty.KustomizeParser, node *KustomizeNode) error {
	k, err := kp.GetKustomization(fSys, node.Path)
	if err != nil {
		return err
	}

	// handle resources
	for _, r := range k.Resources {
		_, err := git.NewRepoSpecFromUrl(r)
		if err == nil {
			handleGitRepo(node, r)
		} else if IsRemoteFile(r) {
			handleRemoteFile(node, r)
		} else {
			_, err := handleLocal(node, r, fSys, kp)
			if err != nil {
				return err
			}
		}
	}

	// handle openapi ref
	if openApiPath, exists := k.OpenAPI["path"]; exists {
		node.AddChild(openApiPath)
	}

	return nil
}

func handleRemoteFile(node *KustomizeNode, path string) *KustomizeNode {
	return node.AddChild(path)
}

func handleGitRepo(node *KustomizeNode, path string) *KustomizeNode {
	return handleRemoteFile(node, path)
}

func handleLocal(node *KustomizeNode, path string, fSys filesys.FileSystem, kp *krusty.KustomizeParser) (*KustomizeNode, error) {
	newPath := path
	if !filepath.IsAbs(path) {
		newPath = filepath.Join(node.Path, path)
	}
	newNode := node.AddChild(newPath)
	if fSys.IsDir(newPath) {
		fmt.Println("doing subdir?", newPath)
		err := parseDir(fSys, kp, newNode)
		return newNode, err
	}

	return newNode, nil
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

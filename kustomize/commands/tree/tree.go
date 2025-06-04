package tree

import (
	"io"
	"os"

	"github.com/spf13/cobra"
)

type Options struct {
	Writer io.Writer
}

func NewCmdTree(w io.Writer) *cobra.Command {
	o := newOptions(w)
	versionCmd := cobra.Command{
		Use:     "tree",
		Example: `kustomize tree`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := o.Validate(args); err != nil {
				return err
			}
			if err := o.Run(); err != nil {
				return err
			}
			return nil
		},
	}

	return &versionCmd
}

func newOptions(w io.Writer) *Options {
	if w == nil {
		w = io.Writer(os.Stdout)
	}
	return &Options{Writer: w}
}

func (o *Options) Validate(_ []string) error {
	return nil
}

func (o *Options) Run() error {
	println("test tree command")
	return nil
}

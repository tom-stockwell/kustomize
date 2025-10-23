package ls

import (
	"fmt"

	"github.com/spf13/pflag"
)

type OutputStyle string

const (
	OutputStyleTree OutputStyle = "tree"
	OutputStyleList OutputStyle = "list"
)

const defaultOutputStyle = OutputStyleList

// implement the pflag.Value interface so this can be used as a flag type
func (e *OutputStyle) String() string {
	return string(*e)
}

func (e *OutputStyle) Set(v string) error {
	switch v {
	case string(OutputStyleTree), string(OutputStyleList):
		*e = OutputStyle(v)
	case "":
		*e = defaultOutputStyle
	default:
		return fmt.Errorf(`must be one of the following: "%s", "%s"`, OutputStyleList, OutputStyleTree)
	}
	return nil
}

// Type is only used in help text
func (e *OutputStyle) Type() string {
	return "OutputStyle"
}

func AddFlagStyle(set *pflag.FlagSet) {
	set.VarPF(
		&theFlags.style,
		"style",
		"s",
		fmt.Sprintf(`The output style to use. Defaults to "%s"`, defaultOutputStyle),
	)
}

func init() {
	theFlags.style = defaultOutputStyle
}

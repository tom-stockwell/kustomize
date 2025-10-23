package ls

import (
	"fmt"

	"github.com/xlab/treeprint"
)

func flatTreePrinter(item *treeprint.Node) {
	fmt.Printf("%v\n", item.Value)
}

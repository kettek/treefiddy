package registry

import (
	"io/fs"

	"github.com/kettek/treefiddy/types"
)

type Plugin struct {
	// ConfigFunc         func(map[string]any)
	TreeNodeMangleFunc func(types.FileReference) (types.NodeMangling, error)
	TreeSortFunc       func(a, b fs.FileInfo) int
	TreeFilterFunc     func(a fs.FileInfo) bool
}

package registry

import (
	"io/fs"

	"github.com/kettek/treefiddy/types"
)

type Plugin struct {
	// ConfigFunc         func(map[string]any)
	TreeNodeMangleFunc func(types.FileReference) (string, error)
	TreeSortFunc       func(a, b fs.FileInfo) int
	TreeFilterFunc     func(a fs.FileInfo) bool
}

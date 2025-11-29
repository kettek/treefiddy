package registry

import (
	"io/fs"

	"github.com/kettek/treefiddy/types"
)

type Plugin struct {
	OnInit        func() error
	OnTreeRefresh func() error
	// ConfigFunc         func(map[string]any)
	TreeNodeMangleFunc func(types.FileReference, types.NodeMangling) (types.NodeMangling, error)
	TreeSortFunc       func(a, b fs.FileInfo) int
	TreeFilterFunc     func(a fs.FileInfo) bool
	//
}

package registry

import (
	"io/fs"

	"github.com/kettek/treefiddy/types"
)

type (
	OnInitFunc         func() error
	OnTreeRefreshFunc  func() error
	TreeNodeMangleFunc func(types.FileReference, types.NodeMangling) (types.NodeMangling, error)
	TreeSortFunc       func(a, b fs.FileInfo) int
	TreeFilterFunc     func(a fs.FileInfo) bool
)

type Plugin struct {
	OnInit        OnInitFunc
	OnTreeRefresh OnTreeRefreshFunc

	TreeNodeMangle TreeNodeMangleFunc
	TreeSort       TreeSortFunc
	TreeFilter     TreeFilterFunc
}

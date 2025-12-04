package registry

import (
	"time"

	"github.com/kettek/treefiddy/types"
)

type (
	OnInitFunc         func() error
	OnTreeRefreshFunc  func() error
	TreeNodeMangleFunc func(types.FileReference, types.NodeMangling) (types.NodeMangling, error)
	TreeSortFunc       func(a, b types.FileReference) int
	TreeFilterFunc     func(a types.FileReference) bool
	EdictFunc          func(ctx types.EdictContext) types.EdictContext
	Periodic           struct {
		Time time.Duration
		Func func()
	}
)

type Plugin struct {
	OnInit        OnInitFunc
	OnTreeRefresh OnTreeRefreshFunc

	TreeNodeMangle TreeNodeMangleFunc
	TreeSort       TreeSortFunc
	TreeFilter     TreeFilterFunc

	Edicts map[string]EdictFunc

	Periodics []Periodic
}

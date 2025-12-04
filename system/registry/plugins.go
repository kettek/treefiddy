package registry

import "maps"

func RefreshPluginFuncs() {
	PluginOnTreeRefreshFuncs = nil
	PluginTreeNodeMangleFuncs = nil
	PluginTreeSortFuncs = nil
	PluginTreeFilterFuncs = nil
	PluginEdicts = make(map[string]EdictFunc)

	for _, system := range systems {
		for _, plugin := range system.Plugins() {
			if plugin.OnTreeRefresh != nil {
				PluginOnTreeRefreshFuncs = append(PluginOnTreeRefreshFuncs, plugin.OnTreeRefresh)
			}
			if plugin.TreeNodeMangle != nil {
				PluginTreeNodeMangleFuncs = append(PluginTreeNodeMangleFuncs, plugin.TreeNodeMangle)
			}
			if plugin.TreeSort != nil {
				PluginTreeSortFuncs = append(PluginTreeSortFuncs, plugin.TreeSort)
			}
			if plugin.TreeFilter != nil {
				PluginTreeFilterFuncs = append(PluginTreeFilterFuncs, plugin.TreeFilter)
			}
			if plugin.Edicts != nil {
				maps.Copy(PluginEdicts, plugin.Edicts)
			}
		}
	}
}

var (
	PluginOnTreeRefreshFuncs  []OnTreeRefreshFunc
	PluginTreeNodeMangleFuncs []TreeNodeMangleFunc
	PluginTreeSortFuncs       []TreeSortFunc
	PluginTreeFilterFuncs     []TreeFilterFunc
	PluginEdicts              map[string]EdictFunc
)

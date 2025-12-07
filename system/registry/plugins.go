package registry

import (
	"maps"

	"github.com/kettek/treefiddy/types"
)

func RefreshPluginFuncs() {
	PluginOnTreeRefreshFuncs = nil
	PluginTreeNodeMangleFuncs = nil
	PluginTreeSortFuncs = nil
	PluginTreeFilterFuncs = nil
	PluginEdicts = make(map[string]EdictFunc)
	PluginEdictNames = nil
	PluginModes = make(types.Modes)
	PluginBinds = nil

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
				for k := range plugin.Edicts {
					PluginEdictNames = append(PluginEdictNames, k)
				}
			}
			if plugin.Config.Modes != nil {
				maps.Copy(PluginModes, plugin.Config.Modes)
			}
			if plugin.Config.Binds != nil {
				PluginBinds = append(PluginBinds, plugin.Config.Binds...)
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
	PluginEdictNames          []string
	PluginModes               types.Modes
	PluginBinds               types.Binds
)

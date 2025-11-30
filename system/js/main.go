// Package js provides a qjs-backed plugin system for treefiddy.
package js

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"github.com/fastschema/qjs"
	"github.com/kettek/treefiddy/system/registry"
	"github.com/kettek/treefiddy/types"
)

type System struct {
	runtime  *qjs.Runtime
	context  *qjs.Context
	plugins  []Plugin
	rplugins []registry.Plugin
}

type Plugin struct {
	registry.Plugin
	name         string
	path         string
	valuesToFree []*qjs.Value
	permissions  pluginPermissions
}

type pluginPermissions struct {
	exec []string
}

func (s *System) Name() string {
	return "js"
}

func (s *System) Init() error {
	rt, err := qjs.New()
	if err != nil {
		return err
	}
	s.runtime = rt
	s.context = rt.Context()

	return nil
}

func (s *System) PopulatePlugins() error {
	systemDir, err := registry.SystemDir(s.Name())
	if err != nil {
		return err
	}

	if entries, err := os.ReadDir(systemDir); err != nil {
		return err
	} else {
		for _, entry := range entries {
			if entry.IsDir() {
				// Look for plugin.js
				if _, err := os.Stat(filepath.Join(systemDir, entry.Name(), "index.js")); err != nil {
					return err
				}
				s.plugins = append(s.plugins, Plugin{
					name: entry.Name(),
					path: filepath.Join(systemDir, entry.Name(), "index.js"),
				})
			} else if strings.HasSuffix(entry.Name(), ".js") {
				s.plugins = append(s.plugins, Plugin{
					name: entry.Name()[:len(entry.Name())-len(filepath.Ext(entry.Name()))],
					path: filepath.Join(systemDir, entry.Name()),
				})
			}
		}
	}

	return nil
}

func (s *System) PluginNames() (names []string) {
	for _, p := range s.plugins {
		names = append(names, p.name)
	}
	return
}

func (s *System) LoadPlugin(name string) error {
	var plugin *Plugin
	for _, p := range s.plugins {
		if p.name == name {
			plugin = &p
		}
	}

	if plugin == nil {
		return fmt.Errorf("no such plugin %s", name)
	}

	bytes, err := os.ReadFile(plugin.path)
	if err != nil {
		return err
	}

	byteCode, err := s.context.Compile(plugin.path, qjs.Code(string(bytes)), qjs.TypeModule())
	if err != nil {
		return err
	}

	val, err := s.context.Eval(plugin.path, qjs.Bytecode(byteCode), qjs.TypeModule())
	if err != nil {
		return err
	}

	propNames, err := val.GetOwnPropertyNames()
	if err != nil {
		return err
	}

	for _, propName := range propNames {
		switch propName {
		case "permissions":
			perms := val.GetPropertyStr(propName)
			defer perms.Free()
			permNames, err := perms.GetOwnPropertyNames()
			if err != nil {
				return err
			}
			for _, permName := range permNames {
				switch permName {
				case "exec":
					var executables []string
					perm := perms.GetPropertyStr(permName)
					defer perm.Free()
					execArray, err := perm.ToArray()
					if err != nil {
						return err
					}
					for _, execItem := range execArray.GetOwnProperties() {
						execName := execArray.GetPropertyStr(execItem.String())
						defer execName.Free()
						executables = append(executables, execName.String())
					}
					// TODO: Show prompt for permissions to execute the given executables.
					plugin.permissions.exec = executables
				default:
					return fmt.Errorf("unknown permission %s", permName)
				}
			}
		case "exec":
			execFunc, err := qjs.FuncToJS(s.context, func(cmd string, args ...string) (string, error) {
				if !slices.Contains(plugin.permissions.exec, cmd) {
					return "", fmt.Errorf("exec permission not granted for cmd %s", cmd)
				}
				out, err := exec.Command(cmd, args...).Output()
				return string(out), err
			})
			if err != nil {
				return err
			}
			plugin.valuesToFree = append(plugin.valuesToFree, execFunc)
			val.SetPropertyStr(propName, execFunc)
		case "mangleTreeNode":
			mangleFunc := val.GetPropertyStr(propName)
			plugin.valuesToFree = append(plugin.valuesToFree, mangleFunc)
			goMangleFunc, err := qjs.JsFuncToGo[func(types.FileReference, types.NodeMangling) (map[string]any, error)](mangleFunc)
			if err != nil {
				return err
			}

			// TODO: Is it possible to just have qjs return the converted type...? It seems the return value is always `map[string]any` and does not do any type conversions for return values...
			mfn := func(fr types.FileReference, mangling types.NodeMangling) (types.NodeMangling, error) {
				jmangled, err := goMangleFunc(fr, mangling)
				if err != nil {
					return types.NodeMangling{}, err
				}
				return types.NodeMangling{
					Name:        jmangled["Name"].(string),
					Color:       jmangled["Color"].(string),
					Prefix:      jmangled["Prefix"].(string),
					PrefixColor: jmangled["PrefixColor"].(string),
					Suffix:      jmangled["Suffix"].(string),
					SuffixColor: jmangled["SuffixColor"].(string),
				}, err
			}
			plugin.TreeNodeMangle = mfn
		case "sortTreeNodes":
			sortFunc := val.GetPropertyStr(propName)
			plugin.valuesToFree = append(plugin.valuesToFree, sortFunc)
			goSortFunc, err := qjs.JsFuncToGo[func(a, b types.FileReference) int](sortFunc)
			if err != nil {
				return err
			}
			plugin.TreeSort = goSortFunc
		case "filterTreeNode":
			filterFunc := val.GetPropertyStr(propName)
			plugin.valuesToFree = append(plugin.valuesToFree, filterFunc)
			goFilterFunc, err := qjs.JsFuncToGo[func(a fs.FileInfo) bool](filterFunc)
			if err != nil {
				return err
			}
			plugin.TreeFilter = goFilterFunc
		case "onInit":
			fn := val.GetPropertyStr(propName)
			plugin.valuesToFree = append(plugin.valuesToFree, fn)
			goFn, err := qjs.JsFuncToGo[func() error](fn)
			if err != nil {
				return err
			}
			plugin.OnInit = goFn
		case "onTreeRefresh":
			fn := val.GetPropertyStr(propName)
			plugin.valuesToFree = append(plugin.valuesToFree, fn)
			goFn, err := qjs.JsFuncToGo[func() error](fn)
			if err != nil {
				return err
			}
			plugin.OnTreeRefresh = goFn
		}
	}

	s.rplugins = append(s.rplugins, plugin.Plugin)

	return nil
}

func (s *System) unloadPlugin(plugin Plugin) {
	for _, val := range plugin.valuesToFree {
		val.Free()
	}
	plugin.valuesToFree = nil
}

func (s *System) UnloadPlugins() {
	for _, plugin := range s.plugins {
		s.unloadPlugin(plugin)
	}
	s.plugins = nil
	s.rplugins = nil
}

func (s *System) Deinit() error {
	// Always unload 'em.'
	s.UnloadPlugins()

	return nil
}

func (s *System) Plugins() []registry.Plugin {
	return s.rplugins
}

func init() {
	registry.Register(&System{})
}

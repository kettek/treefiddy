// Package js provides a qjs-backed plugin system for treefiddy.
package js

import (
	"errors"
	"fmt"
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
	commands registry.Commands // commands interface, should be app
}

type Plugin struct {
	registry.Plugin
	name         string
	path         string
	valuesToFree []*qjs.Value
	permissions  pluginPermissions
}

func (p *Plugin) assignFunc(obj *qjs.Value, name string, jsFunc *qjs.Value) error {
	p.valuesToFree = append(p.valuesToFree, jsFunc)
	obj.SetPropertyStr(name, jsFunc)
	return nil
}

type pluginPermissions struct {
	exec []string
}

func (s *System) Name() string {
	return "js"
}

func (s *System) Init(commands registry.Commands) error {
	rt, err := qjs.New()
	if err != nil {
		return err
	}
	s.runtime = rt
	s.context = rt.Context()
	s.commands = commands

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
		case "edicts":
			edicts := val.GetPropertyStr(propName)
			defer edicts.Free()
			edictNames, err := edicts.GetOwnPropertyNames()
			if err != nil {
				return err
			}
			plugin.Edicts = make(map[string]registry.EdictFunc)
			for _, edictName := range edictNames {
				edictFunc := edicts.GetPropertyStr(edictName)
				plugin.valuesToFree = append(plugin.valuesToFree, edictFunc)

				goEdictFunc, err := qjs.JsFuncToGo[func(types.EdictContext) (map[string]any, error)](edictFunc)
				if err != nil {
					return err
				}
				efn := func(ctx types.EdictContext) types.EdictContext {
					res, err := goEdictFunc(ctx)
					if err != nil {
						ctx.Err = err
						return ctx
					}
					var ctx2 types.EdictContext
					for k, v := range res {
						switch k {
						case "Root":
							ctx2.Root = v.(string)
						case "Selected":
							ctx2.Selected = v.(string)
						case "Arguments":
							for _, v2 := range res["Arguments"].([]any) {
								ctx2.Arguments = append(ctx2.Arguments, v2.(string))
							}
						case "Err":
							if v != nil {
								ctx2.Err = errors.New(v.(string))
							}
						case "Msg":
							ctx2.Msg = v.(string)
						}
					}
					ctx2.Previous = &ctx
					return ctx2
				}
				plugin.Edicts[edictName] = efn
			}
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
		case "popup":
			jsFunc, err := qjs.FuncToJS(s.context, func(v string) error {
				s.commands.Popup(v)
				return nil
			})
			if err != nil {
				return err
			}
			plugin.valuesToFree = append(plugin.valuesToFree, jsFunc)
			val.SetPropertyStr(propName, jsFunc)
		case "refreshTree":
			if jsFunc, err := qjs.FuncToJS(s.context, func() error {
				s.commands.RefreshTree()
				return nil
			}); err != nil {
				return err
			} else if err := plugin.assignFunc(val, propName, jsFunc); err != nil {
				return err
			}
		case "focusTree":
			if jsFunc, err := qjs.FuncToJS(s.context, func() error {
				s.commands.FocusTree()
				return nil
			}); err != nil {
				return err
			} else if err := plugin.assignFunc(val, propName, jsFunc); err != nil {
				return err
			}
		case "focusLocation":
			if jsFunc, err := qjs.FuncToJS(s.context, func() error {
				s.commands.FocusTree()
				return nil
			}); err != nil {
				return err
			} else if err := plugin.assignFunc(val, propName, jsFunc); err != nil {
				return err
			}
		case "focusInput":
			if jsFunc, err := qjs.FuncToJS(s.context, func() error {
				s.commands.FocusInput()
				return nil
			}); err != nil {
				return err
			} else if err := plugin.assignFunc(val, propName, jsFunc); err != nil {
				return err
			}
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
		case "sortTreeNode":
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
			goFilterFunc, err := qjs.JsFuncToGo[func(a types.FileReference) bool](filterFunc)
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

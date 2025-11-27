// Package js provides a qjs-backed plugin system for treefiddy.
package js

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/fastschema/qjs"
	"github.com/kettek/treefiddy/system/registry"
)

type System struct {
	runtime *qjs.Runtime
	context *qjs.Context
	plugins []registry.Plugin
}

type Plugin struct {
	Features struct {
		treeNodeMangleFunc *qjs.Value
		TreeNodeMangleFunc func(string, bool) string
	}
}

func (plugin *Plugin) TreeNodeMangleFunc() func(string, bool) string {
	return plugin.Features.TreeNodeMangleFunc
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

func (s *System) LoadPlugins() error {
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
				if _, err := os.Stat(filepath.Join(systemDir, entry.Name(), "plugin.js")); err != nil {
					return err
				}
				if err := s.loadPlugin(filepath.Join(systemDir, entry.Name(), "plugin.js")); err != nil {
					return err
				}
			} else if strings.HasSuffix(entry.Name(), ".js") {
				// Parse it as a stand-alone plugin.
				if err := s.loadPlugin(filepath.Join(systemDir, entry.Name())); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (s *System) loadPlugin(path string) error {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	byteCode, err := s.context.Compile(path, qjs.Code(string(bytes)), qjs.TypeModule())
	if err != nil {
		return err
	}

	val, err := s.context.Eval(path, qjs.Bytecode(byteCode), qjs.TypeModule())
	if err != nil {
		return err
	}

	propNames, err := val.GetOwnPropertyNames()
	if err != nil {
		return err
	}

	var plugin Plugin

	for _, propName := range propNames {
		switch propName {
		case "mangleTreeNode":
			mangleFunc := val.GetPropertyStr(propName)
			goMangleFunc, err := qjs.JsFuncToGo[func(string, bool) string](mangleFunc)
			if err != nil {
				return err
			}
			plugin.Features.treeNodeMangleFunc = mangleFunc
			plugin.Features.TreeNodeMangleFunc = goMangleFunc
		}
	}

	s.plugins = append(s.plugins, &plugin)

	return nil
}

func (s *System) unloadPlugin(p registry.Plugin) {
	plugin := p.(*Plugin)
	if plugin.Features.TreeNodeMangleFunc != nil {
		plugin.Features.treeNodeMangleFunc.Free()
	}
}

func (s *System) UnloadPlugins() {
	for _, plugin := range s.plugins {
		s.unloadPlugin(plugin)
	}
	s.plugins = nil
}

func (s *System) Deinit() error {
	// Always unload 'em.'
	s.UnloadPlugins()

	return nil
}

func (s *System) Plugins() []registry.Plugin {
	return s.plugins
}

func init() {
	registry.Register(&System{})
}

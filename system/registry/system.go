package registry

type System interface {
	Name() string
	Init(Commands) error
	Deinit() error
	PopulatePlugins() error
	PluginNames() []string
	LoadPlugin(name string) error
	Plugins() []Plugin
}

var systems []System

func Systems() []System {
	return systems
}

package registry

type System interface {
	Name() string
	Init(Commands) error
	Deinit() error
	PopulatePlugins() error
	PluginNames() []string
	LoadPlugin(name string) error
	Plugins() []Plugin
	WritePluginConfig(name string) error
}

var systems []System

func Systems() []System {
	return systems
}

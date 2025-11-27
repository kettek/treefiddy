package registry

type System interface {
	Name() string
	Init() error
	Deinit() error
	LoadPlugins() error
	Plugins() []Plugin
}

var systems []System

func Systems() []System {
	return systems
}

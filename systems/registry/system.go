package registry

type System interface {
	Name() string
	Init() error
	Deinit() error
	LoadPlugins() error
	Plugins() []Plugin
	// RegisterPlugin("name", TYPE)
	// UnregisterPlugin("name")
}

type Plugin interface {
	TreeNodeMangleFunc() func(string, bool) string
}

var systems []System

func Systems() []System {
	return systems
}

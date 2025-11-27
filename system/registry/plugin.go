package registry

type Plugin struct {
	TreeNodeMangleFunc func(string, bool) string
}

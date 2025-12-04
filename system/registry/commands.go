package registry

type Commands interface {
	RefreshTree()
	FocusTree()
	FocusLocation()
	FocusInput()
	Popup(string)
}

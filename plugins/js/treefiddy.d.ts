export interface Plugin {
// settings
	// permissions indicate to the host what permissions the plugin seeks.
	permissions?: {
		// exec is a list of operating system commands the plugin wishes to use. These _must_ be specified if the exec function is to be used. If an item does not exist in this list or the user denies the permissions, then exec will not be able to call that item.
		exec?: string[],
	},

// Plugin->Host function hooks. These are stubs that must exist on the plugin for the host to populate them with the appropriate callback functions.
	// exec calls an arbitrary command on the host system. Permissions to run specific commands must be specified in the permissions.exec plugin field.
	exec?: (cmd: string, ...args: string[]) => string,
	// refreshTree refreshes the tree.
	refreshTree?: func() => void,
	// focusTree focuses the file tree.
	focusTree?: func() => void,
	// focusLocation focuses the location field.
	focusLocation?: func() => void,
	// focusInput focuses the input/status field.
	focusInput?: func() => void,

	periodics?: PeriodicFunc[],

// Host->Plugin event hooks. These are event handlers functions for various host events.
	// onInit is called when the plugin is first initialized. This is before any tree state exists.
	onInit?: () => void,
	// onTreeRefresh is called when the tree is refreshed.
	onTreeRefresh?: () => void,

// Host->Plugin function hooks. These are functions that the host calls to do various things such as sorting or filtering.
	// mangleTreeNode modifies the passed in mangled object and returns a new mangled object to represent a file or directory.
	mangleTreeNode?: (node: {Name: string, Path: string; Dir: boolean}, mangled: Mangled) => Mangled
	// sortTreeNode is a sorting function that is called on each file or directory.
	sortTreeNode?: (a: FileReference, b: FileReference) => number
	// filterTreeNode is a filtering function that is called on each file or directory.
	filterTreeNode?: (a: FileReference) => boolean
}

// FileReference is the lower-level type for a file node's state.
export type FileReference = {
	OriginalName: string
	Name:         string
	Path:         string
	Dir:          boolean
}

// Mangled is a somewhat poorly named type that controls the rendering of tree nodes.
export type Mangled = {
	Name: string
	Color: string
	Prefix: string
	PrefixColor: string
	Suffix: string
	SuffixColor: string
}

export type PeriodicFunc = {
	delta: string
	func: () => void
}

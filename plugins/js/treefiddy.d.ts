export interface Plugin {
// settings

	// config is the configuration for a plugin. Defaults may be defined by populating this field. This field is merged with values defined in the plugin's config.yaml file.
	config?: Config,
		 
	// permissions indicate to the host what permissions the plugin seeks.
	permissions?: {
		// exec is a list of operating system commands the plugin wishes to use. These _must_ be specified if the exec function is to be used. If an item does not exist in this list or the user denies the permissions, then exec will not be able to call that item.
		exec?: string[],
	},

	edicts?: Record<string, (ctx: EdictContext) => EdictContext>,

// Plugin->Host function hooks. These are stubs that must exist on the plugin for the host to populate them with the appropriate callback functions.
	// exec calls an arbitrary command on the host system. Permissions to run specific commands must be specified in the permissions.exec plugin field.
	exec?: (cmd: string, ...args: string[]) => string,
	// popup causes a popup to appear with the given message text.
	popup?: (msg: string) => void,
	// refreshTree refreshes the tree.
	refreshTree?: () => void,
	// focusTree focuses the file tree.
	focusTree?: () => void,
	// focusLocation focuses the location field.
	focusLocation?: () => void,
	// focusInput focuses the input/status field.
	focusInput?: () => void,

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

export type EdictContext = {
	Root:      string
	Selected:  string
	Arguments: string[]
	Err:      string
	Msg: string
	Previous: EdictContext
}

export type Bind = {
	edict: string
	arguments?: string[]
	rune?: string
	key?: number
}

export type Modes = Record<string, Mode>

export type Mode = {
	rune: string
	binds: Bind[]
}

export type Config = {
		binds?: Record<string, any>,
		modes?: Modes,
}

export type PeriodicFunc = {
	delta: string
	func: () => void
}

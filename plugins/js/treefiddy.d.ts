export type FileReference = {
	OriginalName: string
	Name:         string
	Path:         string
	Dir:          boolean
}

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

// Plugin->Host function hooks
export type ExecHook = (cmd: string, ...args: string[]) => string

// Host->Plugin function hooks
export type MangleTreeNode = (node: {Name: string, Path: string; Dir: boolean}, mangled: Mangled) => Mangled

export type SortTreeNode = (a: FileReference, b: FileReference) => number
export type FilterTreeNode = (a: FileReference) => boolean

// Host->Plugin event hooks
export type OnInit = () => void
export type OnTreeRefresh = () => void

export interface Plugin {
	// settings
	permissions?: {
		exec?: string[],
	},
	periodics?: PeriodicFunc[],
	exec?: ExecHook,

	onInit?: OnInit,
	onTreeRefresh?: OnTreeRefresh,

	mangleTreeNode?: MangleTreeNode
	sortTreeNode?: SortTreeNode
	filterTreeNode?: FilterTreeNode
}

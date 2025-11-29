export type Entry = {
	IsDir: () => boolean;
	Name: () => string;
	Path: () => string;
}

export type Mangled = {
	Name: string
	Color: string
	Prefix: string
	PrefixColor: string
	Suffix: string
	SuffixColor: string
}

// Plugin->Host function hooks
export type ExecHook = (cmd: string, ...args: string[]) => string

// Host->Plugin function hooks
export type MangleTreeNode = (node: {Name: string, Path: string; Dir: boolean}, mangled: Mangled) => Mangled

export type SortTreeNodes = (a: Entry, b: Entry) => number
export type FilterTreeNode = (a: Entry) => boolean

// Host->Plugin event hooks
export type OnInit = () => void
export type OnTreeRefresh = () => void

export interface Plugin {
	// settings
	permissions?: {
		exec?: string[],
	},
	exec?: ExecHook,

	onInit?: OnInit,
	onTreeRefresh?: OnTreeRefresh,

	mangleTreeNode?: MangleTreeNode
	sortTreeNodes?: SortTreeNodes
	filterTreeNode?: FilterTreeNode
}

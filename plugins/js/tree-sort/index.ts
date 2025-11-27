type Entry = {
	IsDir: () => boolean;
	Name: () => string;
	Path: () => string;
}

// TODO: Make this stuff configurable...
export default {
	sortTreeNodes: function(a: Entry, b: Entry): int {
		if (a.IsDir() && !b.IsDir()) {
			return -1
		}
		const an = a.Name().toLowerCase()
		const bn = b.Name().toLowerCase()
		return an.localeCompare(bn)
	},
	filterTreeNode: function(a: Entry): boolean {
		/*if (a.Name() == "node_modules") {
			return false
		}*/
		return true
	},
}


import type { Plugin, Entry, FileReference } from '../treefiddy'

const plugin: Plugin = {
	sortTreeNodes: function(a: FileReference, b: FileReference): number {
		if (a.Dir && !b.Dir) {
			return -1
		} else if (!a.Dir && b.Dir) {
			return 1
		}
		const an = a.Name.toLowerCase()
		const bn = b.Name.toLowerCase()
		return an.localeCompare(bn)
	},
	filterTreeNode: function(a: Entry): boolean {
		/*if (a.Name() == "node_modules") {
			return false
		}*/
		return true
	},
}

export default plugin

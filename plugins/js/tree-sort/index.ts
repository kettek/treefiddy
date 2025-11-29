import type { Plugin, Entry } from '../treefiddy'

const plugin: Plugin = {
	sortTreeNodes: function(a: Entry, b: Entry): number {
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

export default plugin

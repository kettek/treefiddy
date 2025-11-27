import mappings from './mappings.ts'

export default {
	mangleTreeNode: function (node: {Name: string, Path: string; Dir: boolean}): string {
		if (node.Dir) {
			return mappings.dir + " " + node.Name
		}

		let target = node.Name.toLowerCase()

		let match: string
		if (match = mappings.exact[target]) {
			return match + " " + node.Name
		}
		if (match = target.substring(target.lastIndexOf('.'))) {
			match = match.substring(1)
			let res = mappings.extensions[match]
			if (res) {
				return res + " " + node.Name
			}
		}

		return mappings.default + " " + node.Name
	},
}


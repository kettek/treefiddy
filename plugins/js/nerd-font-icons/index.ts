import mappings from './mappings.json'

export default {
	mangleTreeNode: function (node: {Name: string, Path: string; Dir: boolean}): {Name: string, Color: string, Prefix: string, Suffix: string} {
		const mangled: {Name: string, Color: string, Prefix: string, Suffix: string} = {Name: node.Name, Color: "", Suffix: "", Prefix: ""}
		if (node.Dir) {
			mangled.Prefix = mappings.other.dir[0] + " "
			mangled.Color = mappings.other.dir[1]
			return mangled
		}

		let target = node.Name.toLowerCase()

		let match: string
		let res: string
		if (match = mappings.filename[target]) {
				mangled.Prefix = match[0] + " "
				mangled.Color = match[1]
				return mangled
		}
		if (match = target.substring(target.lastIndexOf('.'))) {
			match = match.substring(1)
			if (res = mappings.extensions[match]) {
				mangled.Prefix = res[0] + " "
				mangled.Color = res[1]
				return mangled
			}
		}

		mangled.Prefix = mappings.other.default[0] + " "
		mangled.Color = mappings.other.default[1]

		return mangled
	},
}


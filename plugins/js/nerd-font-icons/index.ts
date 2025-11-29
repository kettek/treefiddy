import type { Plugin, Mangled } from '../treefiddy'
import mappings from './mappings.json'

const plugin: Plugin = {
	mangleTreeNode: function (node: {Name: string, Path: string; Dir: boolean}, mangled: Mangled): Mangled {
		if (node.Dir) {
			mangled.Prefix = mappings.other.dir[0] + " "
			mangled.PrefixColor = mappings.other.dir[1]
			return mangled
		}

		let target = node.Name.toLowerCase()

		let match: string
		let res: string
		if (match = mappings.filename[target]) {
				mangled.Prefix = match[0] + " "
				mangled.PrefixColor = match[1]
				return mangled
		}
		if (match = target.substring(target.lastIndexOf('.'))) {
			match = match.substring(1)
			if (res = mappings.extensions[match]) {
				mangled.Prefix = res[0] + " "
				mangled.PrefixColor = res[1]
				return mangled
			}
		}

		mangled.Prefix = mappings.other.default[0] + " "
		mangled.PrefixColor = mappings.other.default[1]

		return mangled
	},
}

export default plugin

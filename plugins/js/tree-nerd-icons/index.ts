import type { Mangled, Plugin } from '../treefiddy'
import mappings_ from './mappings.json'

type Mapping = [string, string, string, string]

interface Mappings {
	other: {
		default: Mapping
		dir: Mapping
		sym: Mapping
	}
	filename: Record<string, Mapping>
	extensions: Record<string, Mapping>
}

const mappings = mappings_ as any as Mappings

const plugin: Plugin = {
	mangleTreeNode: function (node: { Name: string; Path: string; Dir: boolean }, mangled: Mangled): Mangled {
		if (node.Dir) {
			mangled.Prefix = mappings.other.dir[0] + ' '
			mangled.PrefixColor = mappings.other.dir[1]
			return mangled
		}

		let target = node.Name.toLowerCase()

		let match: Mapping | undefined
		let res: Mapping | undefined
		if (match = mappings.filename[target]) {
			mangled.Prefix = match[0] + ' '
			mangled.PrefixColor = match[1]
			return mangled
		}
		let ext: string
		if (ext = target.substring(target.lastIndexOf('.'))) {
			ext = ext.substring(1)
			if (res = mappings.extensions[ext]) {
				mangled.Prefix = res[0] + ' '
				mangled.PrefixColor = res[1]
				return mangled
			}
		}

		mangled.Prefix = mappings.other.default[0] + ' '
		mangled.PrefixColor = mappings.other.default[1]

		return mangled
	},
}

export default plugin

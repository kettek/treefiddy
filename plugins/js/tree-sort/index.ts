import type { Config, FileReference, Plugin } from '../treefiddy'

interface LocalPlugin {
	config: Config & {
		directoriesFirst: boolean
	}
}

const plugin: Plugin & LocalPlugin = {
	config: {
		directoriesFirst: true,
	},
	sortTreeNode: function (a: FileReference, b: FileReference): number {
		if (plugin.config.directoriesFirst) {
			if (a.Dir && !b.Dir) {
				return -1
			} else if (!a.Dir && b.Dir) {
				return 1
			}
		}
		const an = a.OriginalName.toLowerCase()
		const bn = b.OriginalName.toLowerCase()
		return an.localeCompare(bn)
	},
}

export default plugin

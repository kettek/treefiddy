import type { Config, FileReference, Plugin } from '../treefiddy'

interface LocalPlugin {
	config: Config & {
		filterFiles: string[],
	}
}

const plugin: Plugin & LocalPlugin = {
	config: {
		filterFiles: [],
	},
	filterTreeNode: function (file: FileReference): boolean {
		if (plugin.config.filterFiles.includes(file.OriginalName)) {
			return false
		}
		return true
	},
}

export default plugin

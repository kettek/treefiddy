import type { Config, Mangled, Plugin } from '../treefiddy'

interface LocalPlugin {
	adjustColor: (path: string) => string
	refreshFiles: () => void
	config: Config & {
		modifiedColor: string
		untrackedColor: string
		stagedSuffix: string
		stagedSuffixColor: string
	}
}

let modifiedPaths: Set<string> = new Set()
let untrackedPaths: Set<string> = new Set()
let stagedPaths: Set<string> = new Set()

function dirname(path: string) {
	return path.substring(0, path.lastIndexOf('/'))
}

const plugin: Plugin & LocalPlugin = {
	config: {
		modifiedColor: 'yellow',
		untrackedColor: 'green',
		stagedSuffix: ' +',
		stagedSuffixColor: 'gray',
	},
	permissions: {
		exec: ['git'],
	},
	exec: (_cmd: string, ..._args: string[]): string => {
		return ''
	},
	refreshFiles: function () {
		modifiedPaths.clear()
		untrackedPaths.clear()
		stagedPaths.clear()

		const modifiedFiles = plugin.exec?.('git', 'ls-files', '-m', '--exclude-standard').split(
			'\n',
		) ?? []
		const untrackedFiles = plugin.exec?.('git', 'ls-files', '-o', '--exclude-standard').split(
			'\n',
		) ?? []
		const stagedFiles = plugin.exec?.('git', 'diff', '--name-only', '--cached')
			.split(
			'\n',
		) ?? []

		for (let file of modifiedFiles) {
			modifiedPaths.add(file)
			for (let dname = dirname(file); dname != ''; dname = dirname(dname)) {
				modifiedPaths.add(dname)
			}
		}

		for (let file of untrackedFiles) {
			untrackedPaths.add(file)
			for (let dname = dirname(file); dname != ''; dname = dirname(dname)) {
				untrackedPaths.add(dname)
			}
		}

		for (let file of stagedFiles) {
			stagedPaths.add(file)
		}
	},
	onTreeRefresh: function () {
		plugin.refreshFiles()
	},
	mangleTreeNode: function (
		node: { Name: string; Path: string; Dir: boolean },
		mangled: Mangled,
	): Mangled {
		mangled.Color = plugin.adjustColor(node.Path)
		if (stagedPaths.has(node.Path)) {
			mangled.Suffix = plugin.config.stagedSuffix
			mangled.SuffixColor = plugin.config.stagedSuffixColor
		}
		return mangled
	},
	adjustColor: function (path: string): string {
		if (modifiedPaths.has(path)) {
			return plugin.config.modifiedColor
		} else if (untrackedPaths.has(path)) {
			return plugin.config.untrackedColor
		}
		return ''
	},
}

export default plugin

import path from 'path'
import type { Plugin, Mangled } from '../treefiddy'

interface LocalPlugin {
	adjustColor: (path: string) => string
	refreshFiles: () => void
}

let modifiedPaths: Set<string> = new Set()
let untrackedPaths: Set<string> = new Set()

const plugin: (Plugin & LocalPlugin) = {
	permissions: {
		exec: ["git"],
	},
	exec: (cmd: string, ...args: string[]): string => {return ""},
	refreshFiles: function() {
		modifiedPaths.clear()
		untrackedPaths.clear()

		const modifiedFiles = plugin.exec?.("git", "ls-files", "-m", "--exclude-standard").split("\n") ?? []
		const untrackedFiles = plugin.exec?.("git", "ls-files", "-o", "--exclude-standard").split("\n") ?? []

		for (let file of modifiedFiles) {
			modifiedPaths.add(file)
			for (let dirname = path.dirname(file); dirname != "."; dirname = path.dirname(dirname)) {
				modifiedPaths.add(dirname)
			}
		}

		for (let file of untrackedFiles) {
			untrackedPaths.add(file)
			for (let dirname = path.dirname(file); dirname != "."; dirname = path.dirname(dirname)) {
				untrackedPaths.add(dirname)
			}
		}
	},
	onTreeRefresh: function() {
		plugin.refreshFiles()
	},
	mangleTreeNode: function (node: {Name: string, Path: string; Dir: boolean}, mangled: Mangled): Mangled {
		mangled.Color = plugin.adjustColor(node.Path)
		return mangled
	},
	adjustColor: function(path: string): string {
		if (modifiedPaths.has(path)) {
			return "yellow"
		} else if (untrackedPaths.has(path)) {
			return "green"
		}
		return ""
	}
}

export default plugin



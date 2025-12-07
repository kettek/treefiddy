import type { EdictContext, Plugin } from '../treefiddy'

interface LocalPlugin {
}

const plugin: Plugin & LocalPlugin = {
	config: {
		modes: {
			git: {
				rune: 'g',
				binds: [
					{
						edict: 'git',
						rune: 'a',
						arguments: ['stage']
					},
					{
						edict: 'git',
						rune: 'r',
						arguments: ['unstage']
					},
					{
						edict: 'git',
						rune: 'd',
						arguments: ['diff']
					},
				]
			}
		},
	},
	permissions: {
		exec: ['git'],
	},
	exec: (_cmd: string, ..._args: string[]): string => {
		return ''
	},
	popup: (_v: string) => {},
	edicts: {
		"git": (ctx: EdictContext): EdictContext => {
			switch (ctx.Arguments[0]) {
				case 'stage':
				case 'add':
					plugin.exec?.('git', 'add', ctx.Selected)
					break
				case 'unstage':
					plugin.exec?.('git', 'restore', '--staged', ctx.Selected)
					break
				case 'diff':
					plugin.popup?.(plugin.exec?.('git', 'diff', ctx.Selected))
					break
			}
			return ctx
		},
	},
}

export default plugin

import type { EdictContext, Plugin } from '../treefiddy'

interface LocalPlugin {
}

const plugin: Plugin & LocalPlugin = {
	permissions: {
		exec: ['git'],
	},
	exec: (_cmd: string, ..._args: string[]): string => {
		return ''
	},
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
			}
			return ctx
		},
	},
}

export default plugin

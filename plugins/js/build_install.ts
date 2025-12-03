import path from 'node:path'
import { execSync } from 'node:child_process'
import { readdir } from 'node:fs/promises'

const plugins = (await readdir('./', { withFileTypes: true })).filter((dirent) => dirent.isDirectory()).map((dirent) => path.resolve(dirent.name))

for (let plugin of plugins) {
	console.log('🔌 build & install: ', path.basename(plugin))
	process.chdir(plugin)
	execSync('deno task install')
}

import path from 'path'
import { execSync } from 'child_process'
import { readdir } from 'fs/promises'

const plugins =  (await readdir("./", { withFileTypes: true })).filter(dirent => dirent.isDirectory()).map(dirent => path.resolve(dirent.name))

for (let plugin of plugins) {
  console.log("🔌 build & install: ", path.basename(plugin))
  process.chdir(plugin)
  execSync("bun install")
}


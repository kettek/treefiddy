import xdg from 'npm:@folder/xdg'
import path from 'node:path'

const dirs = xdg({ expanded: true, subdir: 'treefiddy' })

let configDir = dirs.config.home
if (process.platform === 'win32') {
  configDir = path.dirname(configDir)
}

console.log(path.join(configDir, 'plugins/js', path.basename(process.cwd())))

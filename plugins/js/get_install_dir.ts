import xdg from '@folder/xdg'
import path from 'path'

const dirs = xdg({expanded: true, subdir: 'treefiddy'})

console.log(path.join(dirs.config.home, "plugins/js", path.basename(process.cwd())))

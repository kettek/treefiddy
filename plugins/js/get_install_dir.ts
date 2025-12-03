import xdg from 'npm:@folder/xdg'
import path from 'node:path'

const dirs = xdg({expanded: true, subdir: 'treefiddy'})

console.log(path.join(dirs.config.home, "plugins/js", path.basename(process.cwd())))

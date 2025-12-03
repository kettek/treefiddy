import fs from 'node:fs'

const exts = await (await fetch('https://raw.githubusercontent.com/nvim-tree/nvim-web-devicons/refs/heads/master/lua/nvim-web-devicons/default/icons_by_file_extension.lua')).text()
const exacts = await (await fetch('https://raw.githubusercontent.com/nvim-tree/nvim-web-devicons/refs/heads/master/lua/nvim-web-devicons/default/icons_by_filename.lua')).text()

let re = /^\s*\["(.*)"\]\s*= { icon = "(.*)", color = "(.*)", cterm_color = "(\d*)",\s*name = "(.*)"/g

let extensionsMap = {}
let exactMap = {}

const getParts = (line) => {
	re.lastIndex = -1
	const found = re.exec(line)

	if (!found || found.length < 6) {
		return []
	}
	const ext = found[1]
	const icon = found[2]
	const color = found[3]
	const cterm = found[4]
	const name = found[5]

	return [ext, icon, color, cterm, name]
}

for (let line of exts.split('\n')) {
	if (line[0] == '}' || line[0] == 'r' || line.length == 0) {
		continue
	}
	const parts = getParts(line)
	if (parts.length == 0) continue
	extensionsMap[parts[0]] = parts.slice(1)
}

for (let line of exacts.split('\n')) {
	if (line[0] == '}' || line[0] == 'r' || line.length == 0) {
		continue
	}
	const parts = getParts(line)
	if (parts.length == 0) continue
	exactMap[parts[0]?.toLowerCase()] = parts.slice(1)
}

const mappings = {
	filename: exactMap,
	extensions: extensionsMap,
	other: {
		default: ['', '#ABB2BF', '', 'Default'],
		dir: ['', '#61AFEF', '', 'Directory'],
		sym: ['', '#56B6C2', '', 'Symlink'],
	},
}

fs.writeFileSync('mappings.json', JSON.stringify(mappings, null, '\t'))

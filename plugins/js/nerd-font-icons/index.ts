import mappings from './mappings.ts'

export default {
  mangleTreeNode: function (str: string, isDir: bool): string {
	  if (isDir) {
		  return mappings.dir + " " + str
	  }

	  let target = str.toLowerCase()

	  let match: string
	  if (match = mappings.exact[target]) {
		  return match + " " + str 
	  }
	  if (match = target.substring(target.lastIndexOf('.'))) {
		match = match.substring(1)
	  	let res = mappings.extensions[match]
	  	if (res) {
	  	        return res + " " + str
	  	}
	  }

	  return mappings.default + " " + str
  },
}


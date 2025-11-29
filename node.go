package main

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"slices"
	//	"strconv"

	//"github.com/gdamore/tcell/v2"
	"github.com/kettek/treefiddy/system/registry"
	"github.com/kettek/treefiddy/types"
	"github.com/rivo/tview"
)

func addDirToTreeNode(target *tview.TreeNode, path string) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		panic(err)
	}

	if filterFunc != nil {
		files = slices.Collect(filter(files, func(s fs.FileInfo) bool {
			return filterFunc(s)
		}))
	}

	// TODO: Collect active plugin funcs into one clean slice...
	for _, system := range registry.Systems() {
		for _, plugin := range system.Plugins() {
			if f := plugin.TreeFilterFunc; f != nil {
				files = slices.Collect(filter(files, f))
			}
		}
	}

	for _, system := range registry.Systems() {
		for _, plugin := range system.Plugins() {
			if sort := plugin.TreeSortFunc; sort != nil {
				slices.SortStableFunc(files, sort)
			}
		}
	}

	for _, file := range files {
		path := filepath.Join(path, file.Name())
		isDir := false
		mode := file.Mode()
		if mode&os.ModeDir == os.ModeDir {
			isDir = true
		}
		// Follow the symlink and see if it is a directory or not.
		if mode&os.ModeSymlink == os.ModeSymlink {
			sym, err := filepath.EvalSymlinks(path)
			if err != nil {
				continue
			}
			// Might as well replace file var with our eval'd one.
			file, err = os.Stat(sym)
			if err != nil {
				continue
			}
			isDir = file.IsDir()
		}

		// Run any manglers...
		fr := types.FileReference{
			Name: file.Name(),
			Path: path,
			Dir:  isDir,
		}
		mangling := types.NodeMangling{
			Name: file.Name(),
		}

		for _, system := range registry.Systems() {
			for _, plugin := range system.Plugins() {
				if mangler := plugin.TreeNodeMangleFunc; mangler != nil {
					mangling, err = mangler(fr, mangling)
					if err != nil {
						// TODO: show some sorta err instead of panicking
						panic(err)
					}
				}
			}
		}

		var name string
		if mangling.Prefix != "" {
			if mangling.PrefixColor != "" {
				name += fmt.Sprintf("[%s]%s[-]", mangling.PrefixColor, mangling.Prefix)
			} else {
				name += mangling.Prefix
			}
		}
		if mangling.Color != "" {
			name += fmt.Sprintf("[%s]%s[-]", mangling.Color, mangling.Name)
		} else {
			name += mangling.Name
		}
		if mangling.Suffix != "" {
			if mangling.SuffixColor != "" {
				name += fmt.Sprintf("[%s]%s[-]", mangling.SuffixColor, mangling.Suffix)
			} else {
				name += mangling.Suffix
			}
		}
		fr.Name = name

		node := tview.NewTreeNode(fr.Name).
			SetReference(fr).
			SetSelectable(true)
		node.SetTextStyle(node.GetTextStyle().Background(0)) // Blank out background... can we set this universally?
		if isDir {
			node.SetTextStyle(node.GetTextStyle().Bold(true))
		}
		target.AddChild(node)
	}
}

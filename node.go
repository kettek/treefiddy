package main

import (
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"slices"

	"github.com/gdamore/tcell/v2"
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

	if sortFunc != nil {
		slices.SortStableFunc(files, sortFunc)
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
		node := tview.NewTreeNode(file.Name()).
			SetReference(nodeRef{
				path: path,
				dir:  isDir,
			}).
			SetSelectable(true)
		if isDir {
			node.SetColor(tcell.ColorPink)
		}
		node.SetTextStyle(node.GetTextStyle().Background(0)) // Blank out background... can we set this universally?
		target.AddChild(node)
	}
}

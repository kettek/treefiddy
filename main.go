package main

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"iter"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"syscall"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var (
	filterFunc func(a fs.FileInfo) bool
	sortFunc   func(a, b fs.FileInfo) int
)

func filter[S any](s []S, fn func(S) bool) iter.Seq[S] {
	return func(yield func(s S) bool) {
		for _, v := range s {
			if fn(v) {
				if !yield(v) {
					return
				}
			}
		}
	}
}

type nodeRef struct {
	path string
	dir  bool
}

func main() {
	dir := "."
	if len(os.Args) > 1 {
		dir = os.Args[1]
	}

	/*sortFunc = func(a, b fs.FileInfo) int {
		if a.IsDir() && !b.IsDir() {
			return -1
		}
		return strings.Compare(a.Name(), b.Name())
	}*/
	filterFunc = func(a fs.FileInfo) bool {
		return a.Name()[0] != '.'
	}

	root := tview.NewTreeNode(dir)

	input := tview.NewInputField()
	loc := tview.NewInputField()
	tree := tview.NewTreeView()

	absdir, _ := filepath.Abs(dir)
	loc.SetText(absdir)
	loc.SetFieldWidth(0)
	loc.SetFieldBackgroundColor(tcell.ColorFuchsia)
	loc.SetFieldTextColor(tcell.ColorBlack)

	input.SetFieldWidth(0)
	input.SetFieldBackgroundColor(tcell.ColorFuchsia)
	input.SetFieldTextColor(tcell.ColorBlack)

	tree.
		SetRoot(root).
		SetCurrentNode(root)

	// Don't show root level of tree as a branch.
	tree.SetTopLevel(1)

	add := func(target *tview.TreeNode, path string) {
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

	edit := func(path string) {
		abs, _ := filepath.Abs(path)
		cmd := exec.Command(os.Getenv("EDITOR"), abs)
		cmd.Env = os.Environ()
		cmd.Stdin = nil
		cmd.Stdout = nil
		cmd.Stderr = nil
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Setsid: true,
		}
		if err := cmd.Start(); err != nil {
			panic(err)
		}
		input.SetText(fmt.Sprintf("%d", cmd.Process.Pid))
	}

	add(root, dir)

	var cnode *tview.TreeNode
	var lastKeyPress, lastMousePress time.Time

	tree.SetSelectedFunc(func(node *tview.TreeNode) {
		reference := node.GetReference()
		if reference == nil {
			return // Selecting the root node does nothing.
		}
		children := node.GetChildren()
		if len(children) == 0 {
			// Load and show files in this directory.
			nr := reference.(nodeRef)
			if nr.dir {
				add(node, nr.path)
			} else {
				// If the selection is from a mouse press, do not immediately edit but rather just select and set our current node to it.
				if !lastKeyPress.After(lastMousePress) && cnode != node {
					cnode = node
					return
				}
				edit(nr.path)
			}
		} else {
			// Collapse if visible, expand if collapsed.
			node.SetExpanded(!node.IsExpanded())
		}
	})

	tree.SetChangedFunc(func(node *tview.TreeNode) {
		// Assign cnode (which is our manually tracked current node) only if the selection was done via keyboard.
		if lastKeyPress.After(lastMousePress) {
			cnode = node
		}
	})

	tree.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		lastKeyPress = event.When()
		return event
	})

	tree.SetMouseCapture(func(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
		lastMousePress = event.When()
		// We handle double-click logic manually, so just turn it into a regular click.
		if action == tview.MouseLeftDoubleClick {
			return tview.MouseLeftClick, event
		}
		return action, event
	})

	grid := tview.NewGrid().SetRows(1, 0, 1).SetColumns(0)
	grid.AddItem(loc, 0, 0, 1, 1, 1, 1, false)
	grid.AddItem(tree, 1, 0, 1, 1, 1, 1, true)
	grid.AddItem(input, 2, 0, 1, 1, 1, 1, false)

	tree.SetBackgroundColor(0)

	app := tview.NewApplication()
	app.SetRoot(grid, true)
	app.EnableMouse(true)

	if err := app.Run(); err != nil {
		panic(err)
	}
}

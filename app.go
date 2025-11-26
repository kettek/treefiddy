package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type app struct {
	*tview.Application
	root     string
	location *tview.InputField
	tree     *tview.TreeView
	rootNode *tview.TreeNode
	cmd      *tview.InputField

	cnode                        *tview.TreeNode
	lastKeyPress, lastMousePress time.Time
}

func newApp() app {
	return app{
		Application: tview.NewApplication(),
		location:    tview.NewInputField(),
		tree:        tview.NewTreeView(),
		rootNode:    tview.NewTreeNode(""),
		cmd:         tview.NewInputField(),
	}
}

func (a *app) setup() {
	a.EnableMouse(true)

	// Styling

	a.tree.SetBackgroundColor(0)
	// Don't show root level of tree as a branch.
	a.tree.SetTopLevel(1)

	a.location.SetFieldWidth(0)
	a.location.SetFieldBackgroundColor(tcell.ColorFuchsia)
	a.location.SetFieldTextColor(tcell.ColorBlack)

	a.cmd.SetFieldWidth(0)
	a.cmd.SetFieldBackgroundColor(tcell.ColorFuchsia)
	a.cmd.SetFieldTextColor(tcell.ColorBlack)

	// Functionality
	a.location.SetDoneFunc(func(key tcell.Key) {
		switch key {
		case tcell.KeyEnter:
			if stat, err := os.Stat(a.location.GetText()); err != nil || !stat.IsDir() {
				a.location.SetText(a.root)
			} else {
				a.setRoot(a.location.GetText())
			}
			a.SetFocus(a.tree)
		case tcell.KeyEscape:
			a.location.SetText(a.root)
		case tcell.KeyTab:
			a.SetFocus(a.tree)
		}
	})

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
		a.cmd.SetText(fmt.Sprintf("%d", cmd.Process.Pid))
	}

	a.tree.SetSelectedFunc(func(node *tview.TreeNode) {
		reference := node.GetReference()
		if reference == nil {
			return // Selecting the root node does nothing.
		}
		children := node.GetChildren()
		if len(children) == 0 {
			// Load and show files in this directory.
			nr := reference.(nodeRef)
			if nr.dir {
				addDirToTreeNode(node, nr.path)
			} else {
				// If the selection is from a mouse press, do not immediately edit but rather just select and set our current node to it.
				if !a.lastKeyPress.After(a.lastMousePress) && a.cnode != node {
					a.cnode = node
					return
				}
				edit(nr.path)
			}
		} else {
			// Collapse if visible, expand if collapsed.
			node.SetExpanded(!node.IsExpanded())
		}
	})

	a.tree.SetChangedFunc(func(node *tview.TreeNode) {
		// Assign cnode (which is our manually tracked current node) only if the selection was done via keyboard.
		if a.lastKeyPress.After(a.lastMousePress) {
			a.cnode = node
		}
	})

	a.tree.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		a.lastKeyPress = event.When()
		return event
	})

	a.tree.SetMouseCapture(func(action tview.MouseAction, event *tcell.EventMouse) (tview.MouseAction, *tcell.EventMouse) {
		a.lastMousePress = event.When()
		// We handle double-click logic manually, so just turn it into a regular click.
		if action == tview.MouseLeftDoubleClick {
			return tview.MouseLeftClick, event
		}
		return action, event
	})

	a.tree.SetRoot(a.rootNode)

	// Global Functionality
	a.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyF5 {
			// Refresh root. TODO: Make this not fully reconstruct the tree, somehow.
			a.setRoot(a.root)
		}
		return event
	})

	// Layout
	grid := tview.NewGrid().SetRows(1, 0, 1).SetColumns(0)
	grid.AddItem(a.location, 0, 0, 1, 1, 1, 1, false)
	grid.AddItem(a.tree, 1, 0, 1, 1, 1, 1, true)
	grid.AddItem(a.cmd, 2, 0, 1, 1, 1, 1, false)

	a.SetRoot(grid, true)
}

func (a *app) setRoot(dir string) {
	a.root = dir

	a.tree.GetRoot().ClearChildren()
	a.tree.GetRoot().SetText(dir)

	addDirToTreeNode(a.rootNode, dir)

	a.tree.SetRoot(a.rootNode).
		SetCurrentNode(a.rootNode)

	absdir, _ := filepath.Abs(dir)
	a.location.SetText(absdir)
}

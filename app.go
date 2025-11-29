package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
	_ "github.com/kettek/treefiddy/system/js"
	"github.com/kettek/treefiddy/system/registry"
	"github.com/kettek/treefiddy/types"
	"github.com/rivo/tview"
)

type app struct {
	*tview.Application
	root        string
	location    *tview.InputField
	tree        *tview.TreeView
	rootNode    *tview.TreeNode
	cmd         *tview.InputField
	cmdIsStatus bool

	config Config

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

func (a *app) setup(dir string) {
	// Grab that config.
	a.config = ensureConfig()

	a.EnableMouse(a.config.UseMouse)

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

	a.tree.SetSelectedFunc(func(node *tview.TreeNode) {
		reference := node.GetReference()
		if reference == nil {
			return // Selecting the root node does nothing.
		}
		children := node.GetChildren()
		if len(children) == 0 {
			// Load and show files in this directory.
			nr := reference.(types.FileReference)
			if nr.Dir {
				a.cnode = node
				addDirToTreeNode(node, nr.Path)
			} else {
				// If the selection is from a mouse press, do not immediately edit but rather just select and set our current node to it.
				if !a.lastKeyPress.After(a.lastMousePress) && a.cnode != node {
					a.cnode = node
					return
				}

				// Get our edict based upon click or enter config.
				edict := "edit"
				if a.lastMousePress.After(a.lastKeyPress) {
					if a.config.Actions.Click != "" {
						edict = a.config.Actions.Click
					}
				} else {
					if a.config.Actions.Enter != "" {
						edict = a.config.Actions.Enter
					}
				}

				a.RunEdict(edict, nr.Path, nil)
			}
		} else {
			// Collapse if visible, expand if collapsed.
			node.SetExpanded(!node.IsExpanded())
			// Update our cnode to match this node.
			a.cnode = node
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
		if event.Key() == tcell.KeyTab {
			a.SetFocus(a.cmd)
			return nil
		} else {
			// Check our binds...
			for _, bind := range a.config.Binds {
				if bind.Edict == "" {
					continue
				}
				if (bind.Rune != rune(0) && bind.Rune == event.Rune()) || (bind.Key != 0 && bind.Key == int(event.Key())) {
					nr := a.cnode.GetReference().(types.FileReference)
					a.RunEdict(bind.Edict, nr.Path, nil)
					return nil
				}
			}
		}
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

	a.cmd.SetDoneFunc(func(key tcell.Key) {
		switch key {
		case tcell.KeyEscape:
			a.ClearStatus()
			a.SetFocus(a.tree)
		case tcell.KeyBacktab:
			a.SetFocus(a.tree)
		case tcell.KeyEnter:
			parts := strings.Split(a.cmd.GetText(), " ")
			if len(parts) == 0 {
				a.SetFocus(a.tree)
				return
			}

			// Use the first argument for edict command but also run through our shortcuts to see if it matches any of those.
			edict := parts[0]
			for _, shortcut := range a.config.Shortcuts {
				if shortcut.Keyword == edict {
					edict = shortcut.Edict
					break
				}
			}

			a.RunEdict(edict, a.cnode.GetReference().(types.FileReference).Path, parts[1:])
			a.SetFocus(a.tree)
		}
	})

	a.cmd.SetFocusFunc(func() {
		if a.cmdIsStatus {
			a.ClearStatus()
		}
	})
	a.cmd.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if a.cmdIsStatus {
			a.ClearStatus()
		}
		return event
	})

	// Global Functionality
	a.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyF5:
			a.setRoot(a.root)
		case tcell.KeyRune:
			if event.Rune() == ':' {
				a.SetFocus(a.cmd)
				return nil
			}
		}
		return event
	})

	// Layout
	grid := tview.NewGrid().SetRows(1, 0, 1).SetColumns(0)
	grid.AddItem(a.location, 0, 0, 1, 1, 1, 1, false)
	grid.AddItem(a.tree, 1, 0, 1, 1, 1, 1, true)
	grid.AddItem(a.cmd, 2, 0, 1, 1, 1, 1, false)

	modal := func(p tview.Primitive) tview.Primitive {
		return tview.NewFlex().
			AddItem(nil, 0, 1, false).
			AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
				AddItem(nil, 0, 1, false).
				AddItem(p, 0, 4, true).
				AddItem(nil, 0, 1, false), 0, 10, true).
			AddItem(nil, 0, 1, false)
	}

	a.tree.SetRoot(a.rootNode)

	info := tview.NewTextView()
	info.SetBorder(true)

	pages := tview.NewPages().
		AddPage("picker", grid, true, true).
		AddPage("modal", modal(info), true, false)

	a.SetRoot(pages, true)

	pages.ShowPage("modal")
	info.SetText("Starting...")

	go func() {
		text := ""
		status := func(v string) {
			text += v
			a.QueueUpdateDraw(func() {
				info.SetText(text)
				info.ScrollToEnd()
			})
		}
		// Plugin shenanigans.
		status("SYSTEMS\n")
		if a.config.Systems.JavaScriptPlugins {
			totalStart := time.Now()
			for _, system := range registry.Systems() {
				systemStart := time.Now()
				status("  " + system.Name() + "\n")
				status("    init... ")
				if err := system.Init(); err != nil {
					status(err.Error())
				} else {
					elapsed := time.Since(systemStart)
					status(fmt.Sprintf("ok (%dms)", elapsed.Milliseconds()))
				}
				status("\n    plugins")
				start := time.Now()

				if err := system.PopulatePlugins(); err != nil {
					system.Deinit()
					status("      " + err.Error())
					continue
				}

				for _, name := range system.PluginNames() {
					status("\n      " + name + "... ")
					start := time.Now()
					if err := system.LoadPlugin(name); err != nil {
						status(err.Error())
						continue
					}
					status(fmt.Sprintf("ok (%dms)", time.Since(start).Milliseconds()))
				}
				// Call init.
				for _, plugin := range system.Plugins() {
					if plugin.OnInit != nil {
						if err := plugin.OnInit(); err != nil {
							status(err.Error())
						}
					}
				}
				status(fmt.Sprintf("\n    (%dms)", time.Since(start).Milliseconds()))

				registry.RefreshPluginFuncs()

				elapsed := time.Since(systemStart)
				status(fmt.Sprintf("\n  (%dms)", elapsed.Milliseconds()))
			}
			totalElapsed := time.Since(totalStart)
			status(fmt.Sprintf("\n%dms TOTAL", totalElapsed.Milliseconds()))
		}
		a.QueueUpdateDraw(func() {
			a.setRoot(dir)
			pages.HidePage("modal")
		})
	}()
}

func (a *app) Status(v string) {
	a.cmd.SetText(v)
	a.cmdIsStatus = true
	a.SetFocus(a.tree)
}

func (a *app) ClearStatus() {
	a.cmd.SetText("")
	a.cmdIsStatus = false
}

func (a *app) RunEdict(edict string, selected string, args []string) (string, error) {
	res, err := RunEdict(edict, EdictContext{
		Root:      a.root,
		Selected:  selected,
		Arguments: args,
	})
	if err != nil {
		a.Status(fmt.Sprintf("error: %s", err.Error()))
		return "", err
	}
	a.Status(fmt.Sprintf("%s %s", edict, res))
	// See if we have an edict to run after this one.
	if nextEdict, ok := a.config.Actions.PostEdictEdicts[edict]; ok {
		// FIXME: This could infinite loop!
		return a.RunEdict(nextEdict, selected, args)
	}

	return res, nil
}

func (a *app) setRoot(dir string) {
	absdir, _ := filepath.Abs(dir)
	if err := os.Chdir(absdir); err != nil {
		panic(err)
	}

	// Refresh any registered.
	for _, fn := range registry.PluginOnTreeRefreshFuncs {
		if err := fn(); err != nil {
			panic(err)
		}
	}

	// TODO: if absdir == old root, try not to reconstruct the whole tree.
	a.root = absdir

	a.tree.GetRoot().ClearChildren()
	a.tree.GetRoot().SetText(".")

	addDirToTreeNode(a.rootNode, ".")

	a.tree.SetRoot(a.rootNode).
		SetCurrentNode(a.rootNode)

	// Set cnode to first child if possible.
	if children := a.rootNode.GetChildren(); children != nil {
		a.cnode = children[0]
	}

	a.location.SetText(absdir)
}

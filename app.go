package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"slices"
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
	pages       *tview.Pages
	info        *tview.TextView
	popup       *tview.Flex
	location    *tview.InputField
	tree        *tview.TreeView
	rootNode    *tview.TreeNode
	cmd         *tview.InputField
	cmdIsStatus bool

	config types.Config

	mode *types.Mode

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
	a.config = types.EnsureConfig()

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
	a.location.SetFocusFunc(func() {
		absdir, _ := filepath.Abs(a.root)
		a.location.SetText(absdir)
	})
	a.location.SetBlurFunc(func() {
		absdir, _ := filepath.Abs(a.root)
		a.location.SetText(filepath.Base(absdir))
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
				a.syncNode(node, nr.Path, true)
				// addDirToTreeNode(node, nr.Path)
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

				a.RunEdict(edict, types.EdictContext{
					Selected: nr.Path,
				})
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

		// Handle modez.
		if a.mode != nil {
			if bind := a.mode.GetBind(event.Rune()); bind != nil && bind.Edict != "" {
				nr := a.cnode.GetReference().(types.FileReference)
				var arguments []string
				arguments = append(arguments, bind.Arguments...)
				a.RunEdict(bind.Edict, types.EdictContext{Selected: nr.Path, Arguments: arguments})
				a.Status(fmt.Sprintf("%s %v %s", bind.Edict, bind.Arguments, nr.Path))
			} else {
				a.ClearStatus()
			}
			a.mode = nil
			return nil
		} else {
			for k, m := range a.config.Modes {
				if rune(m.Rune[0]) == event.Rune() {
					a.mode = &m
					var str string
					for i, v := range m.Binds {
						str += fmt.Sprintf("%s %v", v.Rune, v.Arguments)
						if i != len(m.Binds)-1 {
							str += " | "
						}
					}
					a.Status(fmt.Sprintf("MODE: %s > %s", k, str))
					return nil
				}
			}
		}

		// Truck on normally.
		if event.Key() == tcell.KeyTab {
			a.SetFocus(a.cmd)
			return nil
		} else if event.Key() == tcell.KeyBacktab {
			a.SetFocus(a.location)
			return nil
		} else {
			// Check our binds...
			for _, bind := range a.config.Binds {
				if bind.Edict == "" {
					continue
				}
				if (bind.Rune != "" && rune(bind.Rune[0]) == event.Rune()) || (bind.Key != 0 && bind.Key == int(event.Key())) {
					nr := a.cnode.GetReference().(types.FileReference)
					a.RunEdict(bind.Edict, types.EdictContext{Selected: nr.Path})
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
			var arguments []string
			for _, shortcut := range a.config.Shortcuts {
				if shortcut.Keyword == edict {
					edict = shortcut.Edict
					arguments = append(arguments, shortcut.Arguments...)
					break
				}
			}

			a.RunEdict(edict, types.EdictContext{Selected: a.cnode.GetReference().(types.FileReference).Path, Arguments: append(arguments, parts[1:]...)})
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
			a.refreshRoot()
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

	modal := func(p tview.Primitive) *tview.Flex {
		return tview.NewFlex().
			AddItem(nil, 0, 1, false).
			AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
				AddItem(nil, 0, 1, false).
				AddItem(p, 0, 4, true).
				AddItem(nil, 0, 1, false), 0, 10, true).
			AddItem(nil, 0, 1, false)
	}

	a.tree.SetRoot(a.rootNode)

	a.info = tview.NewTextView()
	a.info.SetBorder(true)

	a.popup = modal(a.info)

	a.pages = tview.NewPages().
		AddPage("picker", grid, true, true).
		AddPage("modal", a.popup, true, false)

	a.SetRoot(a.pages, true)

	a.pages.ShowPage("modal")
	a.info.SetText("Starting...")

	// Let's set up some app-specific edicts.
	RegisterEdict("fold-all", Edict{
		Run: func(ctx types.EdictContext) types.EdictContext {
			a.rootNode.CollapseAll()
			return ctx
		},
	})
	RegisterEdict("quit", Edict{
		Run: func(ctx types.EdictContext) types.EdictContext {
			a.Stop()
			return ctx
		},
	})
	RegisterEdict("plugin", Edict{
		Run: func(ctx types.EdictContext) types.EdictContext {
			if len(ctx.Arguments) < 2 {
				return ctx
			}
			switch ctx.Arguments[0] {
			case "save":
				var err error
				found := false
				for _, system := range registry.Systems() {
					if err = system.WritePluginConfig(ctx.Arguments[1]); err == nil {
						found = true
						break
					}
				}
				if found {
					a.Status(fmt.Sprintf("saved %s", ctx.Arguments[0]))
				} else {
					a.Status(err.Error())
				}
			}
			return ctx
		},
	})

	go func() {
		text := ""
		status := func(v string) {
			text += v
			a.QueueUpdateDraw(func() {
				a.info.SetText(text)
				a.info.ScrollToEnd()
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
				if err := system.Init(a); err != nil {
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
			a.pages.HidePage("modal")

			a.popup.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
				if event.Key() == tcell.KeyEscape || event.Key() == tcell.KeyTab {
					a.pages.HidePage("modal")
					return nil
				}
				return event
			})
		})
	}()
}

func (a *app) Popup(v string) {
	// This gofunc is being used because I'm lazy.
	go func() {
		a.QueueUpdateDraw(func() {
			a.pages.ShowPage("modal")
			a.info.SetText(v)
			a.info.ScrollToBeginning()
			a.SetFocus(a.popup)
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

func (a *app) RunEdict(edict string, ctx types.EdictContext) (types.EdictContext, error) {
	ctx.Root = a.root // Maybe don't assign this here...
	res, err := RunEdict(edict, ctx)
	if err != nil {
		a.Status(fmt.Sprintf("error: %s", err.Error()))
		return res, err
	}
	a.Status(fmt.Sprintf("%s %s", edict, res.Msg))
	// See if we have an edict to run after this one.
	if nextEdict, ok := a.config.Actions.PostEdictEdicts[edict]; ok {
		// FIXME: This could infinite loop!
		next, err := a.RunEdict(nextEdict, res)
		res.Wrap(next)
		return res, err
	}

	a.refreshRoot() // Update tree on any edict. TODO: Maybe make this only for certain edicts?

	return res, nil
}

// TODO: Make this function more efficient and probably break it up into separate funcs.
func (a *app) syncNode(node *tview.TreeNode, path string, shouldExpand bool) error {
	ref := node.GetReference()
	if ref == nil {
		// New node, let's do stuff.
		fi, err := os.Stat(path)
		if err != nil {
			return err
		}
		isDir := false
		mode := fi.Mode()
		if mode&os.ModeDir == os.ModeDir {
			isDir = true
		}
		// Follow the symlink and see if it is a directory or not.
		if mode&os.ModeSymlink == os.ModeSymlink {
			sym, err := filepath.EvalSymlinks(path)
			if err != nil {
				return err
			}
			// Might as well replace file var with our eval'd one.
			fi, err = os.Stat(sym)
			if err != nil {
				return err
			}
			isDir = fi.IsDir()
		}

		fr := types.FileReference{
			OriginalName: fi.Name(),
			Name:         fi.Name(),
			Path:         path,
			Dir:          isDir,
		}

		node.SetTextStyle(node.GetTextStyle().Background(0)) // Blank out background... can we set this universally?
		if isDir {
			node.SetTextStyle(node.GetTextStyle().Bold(true))
		}
		ref = fr
	}
	// Old node, let's updatie.
	fr := ref.(types.FileReference)
	fr.Name = fr.OriginalName
	fr.Name = a.mangle(fr)
	node.SetReference(fr).
		SetSelectable(true).
		SetText(fr.Name).
		SetExpanded(shouldExpand)
	if fr.Dir {
		if node.IsExpanded() {
			fis, err := ioutil.ReadDir(path)
			if err != nil {
				return err
			}
			var newFiles []types.FileReference
			var existingChildren []*tview.TreeNode
			var removedChildren []*tview.TreeNode

			// Collect removed and still-existing children.
			for _, child := range node.GetChildren() {
				nr := child.GetReference().(types.FileReference)
				found := false
				for _, fi := range fis {
					path2 := filepath.Join(path, fi.Name())
					if path2 == nr.Path {
						found = true
						break
					}
				}
				if !found {
					removedChildren = append(removedChildren, child)
				} else {
					existingChildren = append(existingChildren, child)
				}
			}

			// Collect new file infos.
			for _, fi := range fis {
				path2 := filepath.Join(path, fi.Name())
				found := false
				for _, child := range node.GetChildren() {
					nr := child.GetReference().(types.FileReference)
					if path2 == nr.Path {
						found = true
						break
					}
				}
				if !found {
					newFiles = append(newFiles, types.FileReference{
						OriginalName: fi.Name(),
						Name:         fi.Name(),
						Path:         path2,
						Dir:          fi.IsDir(),
					})
				}
			}

			// Remove removed children.
			for _, child := range removedChildren {
				if a.cnode == child {
					// Uh... try to set it to a "sibling"
					var prevChild *tview.TreeNode
					children := node.GetChildren()
					for i, child2 := range children {
						if child == child2 {
							if prevChild == nil && i < len(children)-1 {
								a.tree.SetCurrentNode(children[i+1])
							} else {
								a.tree.SetCurrentNode(prevChild)
							}
							break
						}
						prevChild = child2
					}
				}
				node.RemoveChild(child)
			}

			// Update existing.
			for _, child := range existingChildren {
				fr := child.GetReference().(types.FileReference)
				a.syncNode(child, fr.Path, child.IsExpanded())
			}

			// Filter new.
			if filterFunc != nil {
				newFiles = slices.Collect(filter(newFiles, func(s types.FileReference) bool {
					return filterFunc(s)
				}))
			}

			for _, fn := range registry.PluginTreeFilterFuncs {
				newFiles = slices.Collect(filter(newFiles, fn))
			}

			// Add new.
			for _, fr := range newFiles {
				// path2 := filepath.Join(path, fi.Name())
				childNode := tview.NewTreeNode("")
				// childNode.SetReference(fr) // It'd be nice if we could assign this here instead of recreating it later...
				a.syncNode(childNode, fr.Path, false)
				node.AddChild(childNode)
			}
		}

		// Ugh... sorting here feels kinda bad, man.
		children := node.GetChildren()
		for _, fn := range registry.PluginTreeSortFuncs {
			slices.SortStableFunc(children, func(a, b *tview.TreeNode) int {
				return fn(a.GetReference().(types.FileReference), b.GetReference().(types.FileReference))
			})
		}
		node.ClearChildren()
		node.SetChildren(children)
	}
	//}
	return nil
}

func (a *app) mangle(fr types.FileReference) string {
	var err error
	mangling := types.NodeMangling{
		Name: fr.OriginalName,
	}

	for _, mangler := range registry.PluginTreeNodeMangleFuncs {
		mangling, err = mangler(fr, mangling)
		if err != nil {
			// TODO: show some sorta err instead of panicking
			panic(err)
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
	return name
}

func (a *app) refreshRoot() {
	// Refresh any registered.
	for _, fn := range registry.PluginOnTreeRefreshFuncs {
		if err := fn(); err != nil {
			a.Status(err.Error())
		}
	}
	a.syncNode(a.tree.GetRoot(), ".", true)
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

	if a.root == absdir {
		a.syncNode(a.tree.GetRoot(), ".", true)
	} else {
		a.root = absdir

		a.tree.GetRoot().ClearChildren()
		a.tree.GetRoot().SetText(".")

		a.syncNode(a.tree.GetRoot(), ".", true)
	}

	if a.tree.GetRoot() == nil {
		a.tree.SetRoot(a.rootNode).
			SetCurrentNode(a.rootNode)
	}

	// Set cnode to first child if possible.
	if children := a.rootNode.GetChildren(); children != nil {
		a.cnode = children[0]
		a.tree.SetCurrentNode(a.cnode)
	}

	a.location.SetText(filepath.Base(absdir))
}

func (a *app) RefreshTree() {
	a.refreshRoot()
}

func (a *app) FocusTree() {
	a.SetFocus(a.tree)
}

func (a *app) FocusLocation() {
	a.SetFocus(a.location)
}

func (a *app) FocusInput() {
	a.SetFocus(a.cmd)
}

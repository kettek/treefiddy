package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hymkor/trash-go"
	"github.com/kettek/treefiddy/internal"
)

type EdictContext struct {
	Root      string
	Selected  string
	Arguments []string
	// Results from last edict chain.
	Err      error
	Msg      string
	Previous *EdictContext
}

func (ctx *EdictContext) Wrap(c EdictContext) {
	ctx.Previous = &EdictContext{
		Root:      c.Root,
		Selected:  c.Selected,
		Arguments: c.Arguments,
		//
		Err:      c.Err,
		Msg:      c.Msg,
		Previous: c.Previous,
	}
}

func (ctx *EdictContext) Last(msg string, err error) EdictContext {
	ctx.Msg = msg
	ctx.Err = err
	return *ctx
}

func (ctx *EdictContext) Error(err error) EdictContext {
	ctx.Err = err
	return *ctx
}

func (ctx *EdictContext) Ok(msg string) EdictContext {
	ctx.Err = nil
	ctx.Msg = msg
	return *ctx
}

func (ctx *EdictContext) TargetAbsPath() (string, error) {
	if len(ctx.Arguments) == 0 { // "<selected>"
		return filepath.Abs(ctx.Selected)
	} else if len(ctx.Arguments) == 1 { // selected -> arg
		return ctx.AbsPathFromRel(ctx.Arguments[0])
	}
	return "", fmt.Errorf("requires 0 or 1 arguments")
}

func (ctx *EdictContext) AbsPathFromRel(path string) (string, error) {
	if path[0] == '/' { // "/some/location" -> "<rootdir>/some/location"
		path = filepath.Join(ctx.Root, path)
	} else { // "some/location" -> "<dir of selected>/some/location"
		path = filepath.Join(filepath.Dir(ctx.Selected), path)
	}

	abs, err := filepath.Abs(path)

	return abs, err
}

func (ctx *EdictContext) RelPathFromAbs(path string) (string, error) {
	return filepath.Rel(ctx.Root, path)
}

func (ctx *EdictContext) FromToAbsPath() (string, string, error) {
	if len(ctx.Arguments) == 0 {
		return "", "", fmt.Errorf("requires a path")
	}
	if len(ctx.Arguments) == 1 { // selected -> arg
		from, _ := filepath.Abs(ctx.Selected)
		to, _ := ctx.AbsPathFromRel(ctx.Arguments[0])
		return from, to, nil
	}
	if len(ctx.Arguments) == 2 { // path1 -> path2
		from, _ := ctx.AbsPathFromRel(ctx.Arguments[0])
		to, _ := ctx.AbsPathFromRel(ctx.Arguments[1])
		return from, to, nil
	}
	return "", "", fmt.Errorf("requires 1 or 2 arguments only")
}

type Edict struct {
	Run func(ctx EdictContext) EdictContext
}

var edicts map[string]Edict

func init() {
	edicts = make(map[string]Edict)
}

func RegisterEdict(name string, e Edict) {
	edicts[name] = e
}

func HasEdict(name string) bool {
	_, ok := edicts[name]
	return ok
}

func RunEdict(name string, ctx EdictContext) (EdictContext, error) {
	if e, ok := edicts[name]; ok {
		ctx = e.Run(ctx)
		return ctx, ctx.Err
	}
	return ctx, fmt.Errorf("missing edict \"%s\"", name)
}

func init() {
	RegisterEdict("edit", Edict{
		Run: func(ctx EdictContext) EdictContext {
			path, err := ctx.TargetAbsPath()
			if err != nil {
				return ctx.Error(err)
			}

			if err := internal.Exec(os.Getenv("EDITOR"), path); err != nil {
				return ctx.Error(err)
			}
			return ctx.Ok(path)
		},
	})
	RegisterEdict("open", Edict{
		Run: func(ctx EdictContext) EdictContext {
			path, err := ctx.TargetAbsPath()
			if err != nil {
				return ctx.Error(err)
			}

			if err := internal.Open(path); err != nil {
				return ctx.Error(err)
			}
			return ctx.Ok(path)
		},
	})
	RegisterEdict("create", Edict{
		Run: func(ctx EdictContext) EdictContext {
			if len(ctx.Arguments) == 0 {
				return ctx.Error(fmt.Errorf("requires a path"))
			}

			path, err := ctx.TargetAbsPath()
			if err != nil {
				return ctx.Error(err)
			}

			if _, err := os.Stat(path); err != nil && !os.IsNotExist(err) {
				return ctx.Error(err)
			}

			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				return ctx.Error(err)
			}

			if fs, err := os.Create(path); err != nil {
				return ctx.Error(err)
			} else {
				fs.Close()
			}

			// Modify our selected and arguments so chaining will apply to the new file.
			lpath, err := ctx.RelPathFromAbs(path)
			if err != nil {
				return ctx.Error(err)
			}
			ctx.Arguments = nil
			ctx.Selected = lpath

			return ctx.Ok(path)
		},
	})
	RegisterEdict("mkdir", Edict{
		Run: func(ctx EdictContext) EdictContext {
			if len(ctx.Arguments) == 0 {
				return ctx.Error(fmt.Errorf("requires a path"))
			}
			var path string
			// check if selected is a dir, and if so, we root ourself to it.
			if fs, err := os.Stat(ctx.Selected); err != nil {
				return ctx.Error(err)
			} else if fs.IsDir() {
				path = filepath.Join(ctx.Selected, strings.Join(ctx.Arguments, " "))
			} else {
				path = filepath.Join(filepath.Dir(ctx.Selected), strings.Join(ctx.Arguments, " "))
			}
			abs, _ := filepath.Abs(path)
			return ctx.Ok(abs)
		},
	})
	RegisterEdict("remove", Edict{
		Run: func(ctx EdictContext) EdictContext {
			path, err := ctx.TargetAbsPath()
			if err != nil {
				return ctx.Error(err)
			}

			err = os.Remove(path)

			return ctx.Last(path, err)
		},
	})
	RegisterEdict("rename", Edict{
		Run: func(ctx EdictContext) EdictContext {
			from, to, err := ctx.FromToAbsPath()
			if err != nil {
				return ctx.Error(err)
			}

			if err := os.Rename(from, to); err != nil {
				return ctx.Error(err)
			}

			return ctx.Ok(from + "->" + to)
		},
	})
	RegisterEdict("trash", Edict{
		Run: func(ctx EdictContext) EdictContext {
			path, err := ctx.TargetAbsPath()
			if err != nil {
				return ctx.Error(err)
			}

			err = trash.Throw(path)

			return ctx.Last(path, err)
		},
	})
}

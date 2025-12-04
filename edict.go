package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hymkor/trash-go"
	"github.com/kettek/treefiddy/internal"
	"github.com/kettek/treefiddy/system/registry"
	"github.com/kettek/treefiddy/types"
)

type Edict struct {
	Run func(ctx types.EdictContext) types.EdictContext
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

func RunEdict(name string, ctx types.EdictContext) (types.EdictContext, error) {
	// Checkum de plugin edicties.
	if e, ok := registry.PluginEdicts[name]; ok {
		ctx = e(ctx)
		return ctx, ctx.Err
	}
	if e, ok := edicts[name]; ok {
		ctx = e.Run(ctx)
		return ctx, ctx.Err
	}
	return ctx, fmt.Errorf("missing edict \"%s\"", name)
}

func init() {
	RegisterEdict("edit", Edict{
		Run: func(ctx types.EdictContext) types.EdictContext {
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
		Run: func(ctx types.EdictContext) types.EdictContext {
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
		Run: func(ctx types.EdictContext) types.EdictContext {
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
		Run: func(ctx types.EdictContext) types.EdictContext {
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
		Run: func(ctx types.EdictContext) types.EdictContext {
			path, err := ctx.TargetAbsPath()
			if err != nil {
				return ctx.Error(err)
			}

			err = os.Remove(path)

			return ctx.Last(path, err)
		},
	})
	RegisterEdict("rename", Edict{
		Run: func(ctx types.EdictContext) types.EdictContext {
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
		Run: func(ctx types.EdictContext) types.EdictContext {
			path, err := ctx.TargetAbsPath()
			if err != nil {
				return ctx.Error(err)
			}

			err = trash.Throw(path)

			return ctx.Last(path, err)
		},
	})
}

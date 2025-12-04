package types

import (
	"fmt"
	"path/filepath"
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

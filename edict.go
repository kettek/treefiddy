package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

type EdictContext struct {
	Root      string
	Selected  string
	Arguments []string
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
	Run func(ctx EdictContext) (string, error)
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

func RunEdict(name string, ctx EdictContext) (string, error) {
	if e, ok := edicts[name]; ok {
		return e.Run(ctx)
	}
	return "", fmt.Errorf("missing edict \"%s\"", name)
}

func init() {
	RegisterEdict("edit", Edict{
		Run: func(ctx EdictContext) (string, error) {
			path, err := ctx.TargetAbsPath()
			if err != nil {
				return "", err
			}

			cmd := exec.Command(os.Getenv("EDITOR"), path)
			cmd.Env = os.Environ()
			cmd.Stdin = nil
			cmd.Stdout = nil
			cmd.Stderr = nil
			cmd.SysProcAttr = &syscall.SysProcAttr{
				Setsid: true,
			}
			if err := cmd.Start(); err != nil {
				return "", err
			}
			return path, nil
		},
	})
	RegisterEdict("open", Edict{
		Run: func(ctx EdictContext) (string, error) {
			path, err := ctx.TargetAbsPath()
			if err != nil {
				return "", err
			}

			program := "xdg-open"
			// First check if "xdg-open" is available.
			_, err = exec.LookPath("xdg-open")
			// Otherwise default to "open".
			if err != nil {
				program = "open"
			}

			if err := exec.Command(program, path).Start(); err != nil {
				return "", err
			}
			return path, nil
		},
	})
	RegisterEdict("create", Edict{
		Run: func(ctx EdictContext) (string, error) {
			if len(ctx.Arguments) == 0 {
				return "", fmt.Errorf("requires a path")
			}

			path, err := ctx.TargetAbsPath()
			if err != nil {
				return "", err
			}

			if _, err := os.Stat(path); err != nil && !os.IsNotExist(err) {
				return "", err
			}

			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				return "", err
			}

			if fs, err := os.Create(path); err != nil {
				return "", err
			} else {
				fs.Close()
			}

			return path, nil
		},
	})
	RegisterEdict("mkdir", Edict{
		Run: func(ctx EdictContext) (string, error) {
			if len(ctx.Arguments) == 0 {
				return "", fmt.Errorf("requires a path")
			}
			var path string
			// check if selected is a dir, and if so, we root ourself to it.
			if fs, err := os.Stat(ctx.Selected); err != nil {
				return "", err
			} else if fs.IsDir() {
				path = filepath.Join(ctx.Selected, strings.Join(ctx.Arguments, " "))
			} else {
				path = filepath.Join(filepath.Dir(ctx.Selected), strings.Join(ctx.Arguments, " "))
			}
			abs, _ := filepath.Abs(path)
			return abs, nil
		},
	})
	RegisterEdict("remove", Edict{
		Run: func(ctx EdictContext) (string, error) {
			path, err := ctx.TargetAbsPath()
			if err != nil {
				return "", err
			}

			err = os.Remove(path)

			return path, err
		},
	})
	RegisterEdict("rename", Edict{
		Run: func(ctx EdictContext) (string, error) {
			from, to, err := ctx.FromToAbsPath()
			if err != nil {
				return "", err
			}

			if err := os.Rename(from, to); err != nil {
				return "", err
			}

			return from + "->" + to, nil
		},
	})
}

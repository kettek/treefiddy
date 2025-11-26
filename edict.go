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
			var path string
			if len(ctx.Arguments) != 0 {
				path = strings.Join(ctx.Arguments, " ")
			} else {
				path = ctx.Selected
			}
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
				return "", err
			}
			return abs, nil
		},
	})
	RegisterEdict("open", Edict{
		Run: func(ctx EdictContext) (string, error) {
			var path string
			if len(ctx.Arguments) != 0 {
				path = strings.Join(ctx.Arguments, " ")
			} else {
				path = ctx.Selected
			}
			abs, _ := filepath.Abs(path)

			program := "xdg-open"
			// First check if "xdg-open" is available.
			_, err := exec.LookPath("xdg-open")
			// Otherwise default to "open".
			if err != nil {
				program = "open"
			}

			if err := exec.Command(program, abs).Start(); err != nil {
				return "", err
			}
			return abs, nil
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
	RegisterEdict("rm", Edict{
		Run: func(ctx EdictContext) (string, error) {
			var path string
			if len(ctx.Arguments) == 0 {
				path = ctx.Selected
			} else {
				path = strings.Join(append([]string{ctx.Root}, ctx.Arguments...), " ")
			}
			return path, nil
		},
	})
}

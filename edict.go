package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

type Edict struct {
	Run func(string, ...string) (string, error)
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

func RunEdict(name string, selected string, args ...string) (string, error) {
	if e, ok := edicts[name]; ok {
		return e.Run(selected, args...)
	}
	return "", fmt.Errorf("missing edict \"%s\"", name)
}

func init() {
	RegisterEdict("edit", Edict{
		Run: func(selected string, v ...string) (string, error) {
			var path string
			if len(v) != 0 {
				path = strings.Join(v, " ")
			} else {
				path = selected
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
	RegisterEdict("mkdir", Edict{
		Run: func(selected string, v ...string) (string, error) {
			if len(v) == 0 {
				return "", fmt.Errorf("requires a path")
			}
			var path string
			// check if selected is a dir, and if so, we root ourself to it.
			if fs, err := os.Stat(selected); err != nil {
				return "", err
			} else if fs.IsDir() {
				path = filepath.Join(selected, strings.Join(v, " "))
			} else {
				path = strings.Join(v, " ")
			}
			abs, _ := filepath.Abs(path)
			return abs, nil
		},
	})
	RegisterEdict("rm", Edict{
		Run: func(selected string, v ...string) (string, error) {
			var path string
			if len(v) == 0 {
				path = selected
			} else {
				path = strings.Join(v, " ")
			}
			return path, nil
		},
	})
}

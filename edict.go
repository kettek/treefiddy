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
	Run func(...string) (string, error)
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

func RunEdict(name string, args ...string) (string, error) {
	if e, ok := edicts[name]; ok {
		return e.Run(args...)
	}
	return "", fmt.Errorf("missing edict \"%s\"", name)
}

func init() {
	RegisterEdict("edit", Edict{
		Run: func(v ...string) (string, error) {
			path := strings.Join(v, " ")
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
}

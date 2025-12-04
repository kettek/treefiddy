//go:build !windows

// Package internal
package internal

import (
	"os"
	"os/exec"
	"syscall"
)

func Exec(command string, args ...string) error {
	cmd := exec.Command(command, args...)
	cmd.Env = os.Environ()
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true,
	}
	return cmd.Start()
}

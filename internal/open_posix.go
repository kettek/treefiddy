//go:build !windows

// Package internal
package internal

import "os/exec"

func Open(path string) error {
	program := "xdg-open"
	// First check if "xdg-open" is available.
	_, err := exec.LookPath("xdg-open")
	// Otherwise default to "open".
	if err != nil {
		program = "open"
	}

	return Exec(program, path)
}

//go:build windows

// Package internal
package internal

func Open(path string) error {
	return Exec("rundll32.exe", "url.dll,FileProtocolHandler", path)
}

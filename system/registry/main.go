// Package registry provides a common place for packages to register themselves and for treefiddy to make plugin queries.
package registry

import (
	"fmt"
	"os"
	"path/filepath"
)

func Register(system System) error {
	for _, s := range systems {
		if s.Name() == system.Name() {
			return fmt.Errorf("plugin %s already registered", system.Name())
		}
	}

	systems = append(systems, system)

	return nil
}

func SystemDir(name string) (string, error) {
	udir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	systemDir := filepath.Join(udir, "treefiddy", "plugins", name)

	if err := os.MkdirAll(systemDir, 0o755); err != nil {
		return systemDir, err
	}
	return systemDir, nil
}

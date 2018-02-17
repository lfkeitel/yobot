// +build linux,cgo darwin,cgo

package plugins

import (
	"fmt"
	"os"
	"path/filepath"
	"plugin"
)

func Load(path string, modules []string) error {
	if len(modules) == 0 {
		return nil
	}

	fmt.Println("Loading modules")

	for _, module := range modules {
		p := filepath.Join(path, module+".so")
		if !fileExists(p) {
			return fmt.Errorf("Module %s not found", module)
		}

		if _, err := plugin.Open(p); err != nil {
			return err
		}
		fmt.Printf("Loaded %s\n", module)
	}
	return nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// +build linux,cgo darwin,cgo

package plugins

import (
	"fmt"
	"os"
	"path/filepath"
	"plugin"
)

func Load(path string) error {
	if path == "" {
		return nil
	}

	fmt.Println("Loading modules")
	return filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if info == nil {
			return nil
		}
		if info.IsDir() || filepath.Ext(path) != ".so" {
			return nil
		}
		if err != nil {
			return err
		}

		fmt.Printf("Loading module %s\n", filepath.Base(path))
		_, err = plugin.Open(path)
		return err
	})
}

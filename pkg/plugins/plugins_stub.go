// +build !linux,!darwin !cgo

package plugins

import "fmt"

const PluginsSupported = false

func Load(path string, modules []string) error {
	if len(modules) == 0 {
		return nil
	}

	fmt.Println("Modules are not supported in this build")
	return nil
}

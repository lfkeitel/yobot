// +build !linux,!darwin !cgo

package plugins

import "fmt"

func Load(path string) error {
	fmt.Println("Modules are not supported in this build")
	return nil
}

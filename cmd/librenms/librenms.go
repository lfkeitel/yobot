package main

import (
	"fmt"

	"github.com/lfkeitel/yobot/librenms"
)

const (
	address   = ""
	authToken = ""
	hostname  = ""
)

func main() {
	c, err := librenms.NewClient(address)
	if err != nil {
		fmt.Println(err)
		return
	}

	c.SkipTLSVerify()

	if err := c.Login(authToken); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Connected")

	dev, err := c.GetDevice(hostname)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("%#v\n", dev)
}

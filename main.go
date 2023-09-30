package main

import (
	"fmt"
	"os"

	"github.com/hashicorp/packer-plugin-sdk/plugin"
	create "github.com/powa458/packer-plugin-wim/post-processor/create"
)

func main() {
	pps := plugin.NewSet()
	pps.RegisterPostProcessor(plugin.DEFAULT_NAME, new(create.PostProcessor))
	err := pps.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

package main

import (
	"fmt"
	"os"

	"github.com/bobcob7/netmon/cmd"
)

func usage(name string) {
	fmt.Printf("Usage: %s [interface names]...\n", name)
}

func main() {
	interfaceNames := os.Args[1:]
	if len(interfaceNames) == 0 {
		usage(os.Args[0])
		return
	}
	cmd.Run(interfaceNames)
}

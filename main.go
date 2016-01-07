package main

import (
	"flag"
	"fmt"
)

func main() {
	var modeFlag string
	var serverFlag string
	var lastClientFlag bool

	flag.StringVar(&modeFlag, "mode", defaultModeFlag, "set to server for server mode or client for client mode")
	flag.StringVar(&serverFlag, "server", defaultServerFlag, "set the server address")
	flag.BoolVar(&lastClientFlag, "last-client", defaultLastClientFlag, "last client")
	flag.BoolVar(&debugFlag, "debug", debugFlag, "debug mode")

	flag.Parse()

	switch modeFlag {
	case "client":
		clientMain(serverFlag, lastClientFlag)
	case "server":
		serverMain(serverFlag)
	default:
		fmt.Print("Unknown mode.")
		flag.Usage()
	}
}

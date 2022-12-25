package cli

import (
	"flag"
	"fmt"
	"runtime"

	"github.com/fantasticake/fantasticoin/explorer"
	"github.com/fantasticake/fantasticoin/rest"
)

func usage() {
	fmt.Printf("Please use the following flags:\n")
	fmt.Printf("-mode: Start a server with a mode: 'rest','html' (default 'rest')\n")
	fmt.Printf("-port: Set port for a server (default 4000)\n\n")
	runtime.Goexit()
}

func Start() {
	mode := flag.String("mode", "rest", "Start a server with a mode: 'rest','html'")
	port := flag.Int("port", 4000, "Set port for a server")
	flag.Parse()

	switch *mode {
	case "rest":
		rest.Start(*port)
	case "html":
		explorer.Start(*port)
	default:
		usage()
	}
}

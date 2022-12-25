package main

import (
	"github.com/fantasticake/fantasticoin/cli"
	"github.com/fantasticake/fantasticoin/db"
)

func main() {
	defer db.Close()
	cli.Start()
}

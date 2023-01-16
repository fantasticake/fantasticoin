package main

import (
	"github.com/fantasticake/simple-coin/cli"
	"github.com/fantasticake/simple-coin/db"
)

func main() {
	defer db.Close()
	cli.Start()
}

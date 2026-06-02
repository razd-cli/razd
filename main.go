package main

import (
	"os"

	"github.com/razd-cli/razd/internal/cli"
)

func main() {
	app := cli.New()
	os.Exit(app.Run(os.Args[1:]))
}
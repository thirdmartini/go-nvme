package main

import (
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

const appName = "nvmectl"

func commandNotFound(c *cli.Context, command string) {
	log.Fatalf("'%s' is not a %s command. See '%s --help'.",
		command, c.App.Name, c.App.Name)
}

func main() {
	app := cli.NewApp()
	app.Name = appName
	app.Usage = "nvmectl"
	app.Version = "(none)"

	app.Flags = globalFlags
	app.Commands = commands
	app.CommandNotFound = commandNotFound

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

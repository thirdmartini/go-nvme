package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/thirdmartini/go-nvme/api"
	"github.com/thirdmartini/go-nvme/cmd/nvmectl/client"
	"github.com/urfave/cli/v2"
)

var (
	apiUrlFlag = cli.StringFlag{
		Name: "api",
		//Aliases: []string{"api"},
		Required: true,
		Value:    "",
		Usage:    "URL of api server",
	}

	ctlFlag = cli.StringFlag{
		Name: "ctl",
		//Aliases: []string{"api"},
		Value: "",
		Usage: "ctl",
	}

	uuidFlag = cli.StringFlag{
		Name: "uuid",
		//Aliases: []string{"uuid"},
		Required: true,
		Value:    "",
		Usage:    "uuid of target",
	}

	nameFlag = cli.StringFlag{
		Name: "name",
		//Aliases: []string{"name"},
		Value: "",
		Usage: "name of target",
	}

	sizeFlag = cli.Int64Flag{
		Name:     "size",
		Required: true,
		//Aliases: []string{"size"},
		Usage: "size of target",
	}
)

var globalFlags = []cli.Flag{
	&apiUrlFlag,
	&ctlFlag,
}

var commands = []*cli.Command{
	{
		Name:        "target",
		Usage:       "target commands",
		Description: "target commands",
		Subcommands: targetCommands,
	},
}

func mustCreateClient(ctx *cli.Context) api.Client {
	if ctx.String(ctlFlag.Name) != "" {
		client := client.NewClient(ctx.String(apiUrlFlag.Name), ctx.String(ctlFlag.Name))
		if client == nil {
			panic("no client")
		}

		return client
	}

	client := api.NewHTTPClient(ctx.String(apiUrlFlag.Name))
	if client == nil {
		panic("no client")
	}

	return client
}

func handleError(err error) error {
	status := api.Status{
		Code:    http.StatusBadRequest,
		Message: err.Error(),
	}

	data, err := json.MarshalIndent(status, "", "  ")
	fmt.Printf(string(data))
	return err
}

package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
)

var targetCommands = []*cli.Command{
	{
		Name:        "list",
		Usage:       "list targets",
		Description: "list targets",
		Flags:       []cli.Flag{},
		Action:      targetList,
	},

	{
		Name:        "get",
		Usage:       "get targets",
		Description: "get targets",
		Flags: []cli.Flag{
			&uuidFlag,
		},
		Action: targetGet,
	},
	{
		Name:        "destroy",
		Usage:       "destroy targets",
		Description: "destroy targets",
		Flags: []cli.Flag{
			&uuidFlag,
		},
		Action: targetDestroy,
	},
	{
		Name:        "create",
		Usage:       "create targets",
		Description: "create targets",
		Flags: []cli.Flag{
			&nameFlag,
			&sizeFlag,
		},
		Action: targetCreate,
	},
}

func targetList(ctx *cli.Context) error {
	client := mustCreateClient(ctx)
	//defer client.Close()

	volumes, err := client.ListVolumes()
	if err != nil {
		return handleError(err)
	}

	data, err := json.MarshalIndent(volumes, "", "  ")
	if err != nil {
		return err
	}

	fmt.Printf(string(data))
	return nil
}

func targetGet(ctx *cli.Context) error {
	client := mustCreateClient(ctx)
	//defer client.Close()

	volumes, err := client.ListVolumes()
	if err != nil {
		return handleError(err)
	}

	for _, v := range volumes {
		if v.UUID == ctx.String(uuidFlag.Name) {
			data, err := json.MarshalIndent(v, "", "  ")
			if err != nil {
				return err
			}

			fmt.Printf(string(data))
			return nil
		}
	}

	return handleError(os.ErrNotExist)
}

func targetCreate(ctx *cli.Context) error {
	client := mustCreateClient(ctx)
	//defer client.Close()

	volumes, err := client.CreateVolume(ctx.String(nameFlag.Name), "", uint64(ctx.Int64(sizeFlag.Name)))
	if err != nil {
		return handleError(err)
	}

	data, err := json.MarshalIndent(volumes, "", "  ")
	if err != nil {
		return handleError(err)
	}

	fmt.Printf(string(data))
	return nil
}

func targetDestroy(ctx *cli.Context) error {
	client := mustCreateClient(ctx)
	//defer client.Close()

	err := client.DeleteVolume(ctx.String(uuidFlag.Name))
	if err != nil {
		return handleError(err)
	}

	fmt.Printf("{}")
	return nil
}

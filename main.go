package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Commands: []*cli.Command{
			PBFTCommand,
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Errorf("%s", err)
	}
}

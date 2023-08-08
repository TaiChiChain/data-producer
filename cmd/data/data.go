package main

import (
	"fmt"
	"os"
	"time"

	"github.com/urfave/cli/v2"
)

const DefaultParallel = 8
const DefaultQuantity = 2000000

func main() {
	app := &cli.App{
		Name:     "data",
		Usage:    "produce stress test data",
		Compiled: time.Now(),
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "url",
				Usage: "JSON RPC url",
				Value: "http://localhost:8881",
			},
		},
	}

	app.Commands = cli.Commands{
		initCMD,
		generateCMD,
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
	}
}

package main

import (
	"os"

	cli "github.com/urfave/cli/v2"
	teapp "github.com/yzimhao/trading_engine/app"
	"github.com/yzimhao/utilgo/pack"
)

func main() {
	app := &cli.App{
		Name:  "trading_engine",
		Usage: "",
		Flags: []cli.Flag{
			&cli.StringFlag{Name: "config", Aliases: []string{"c"}, Value: "./config.toml", Usage: "config file"},
		},
		Action: func(c *cli.Context) error {
			teapp.Start(c.String("config"))
			return nil
		},
		Commands: []*cli.Command{
			{
				Name:    "version",
				Aliases: []string{"v"},
				Usage:   "print version",
				Action: func(ctx *cli.Context) error {
					pack.ShowVersion()
					return nil
				},
			},
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		panic(err)
	}
}

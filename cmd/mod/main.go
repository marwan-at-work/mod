package main

import (
	"os"

	"github.com/urfave/cli/v2"

	"github.com/marwan-at-work/mod/major"
	"github.com/marwan-at-work/mod/migrate"
)

func main() {
	app := &cli.App{
		Name:  "mod",
		Usage: "upgrade/downgrade semantic import versioning",
		Commands: []*cli.Command{
			{
				Name:        "upgrade",
				Usage:       "mod upgrade [-t]",
				Description: "upgrade go.mod and imports one major or through -t",
				Action:      withExit(upgrade),
				Flags: []cli.Flag{
					&cli.IntFlag{
						Name:    "tag",
						Aliases: []string{"t"},
						Value:   0,
					},
					&cli.StringFlag{
						Name:  "mod-name",
						Value: "",
						Usage: "upgrade a dependency instead of your main module",
					},
				},
			},
			{
				Name:        "downgrade",
				Usage:       "mod downgrade",
				Description: "downgrade go.mod and imports one major version",
				Action:      withExit(downgrade),
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "mod-name",
						Value: "",
						Usage: "downgrade a dependency instead of your main module",
					},
				},
			},
			{
				Name:        "migrate-deps",
				Usage:       "mod migrate-deps -token=<github-token>",
				Description: "migrate your +incompatiable dependencies to Go Modules",
				Action:      withExit(migrateDeps),
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name: "token",
					},
					&cli.IntFlag{
						Name:    "limit",
						Aliases: []string{"l"},
						Value:   -1,
					},
					&cli.BoolFlag{
						Name:  "test",
						Value: false,
					},
				},
			},
		},
	}

	app.Run(os.Args)
}

func upgrade(c *cli.Context) error {
	return major.Run(".", "upgrade", c.String("mod-name"), c.Int("tag"))
}

func downgrade(c *cli.Context) error {
	return major.Run(".", "downgrade", c.String("mod-name"), 0)
}

func migrateDeps(c *cli.Context) error {
	return migrate.Run(c.String("token"), c.Int("limit"), c.Bool("test"))
}

func withExit(f cli.ActionFunc) cli.ActionFunc {
	return func(c *cli.Context) error {
		return handleErr(f(c))
	}
}

func handleErr(err error) error {
	if err != nil {
		return cli.Exit(err, 1)
	}

	return nil
}

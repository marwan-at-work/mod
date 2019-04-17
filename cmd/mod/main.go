package main

import (
	"os"

	"github.com/marwan-at-work/mod/major"
	"github.com/marwan-at-work/mod/migrate"
	"gopkg.in/urfave/cli.v2"
)

func main() {
	buildFlags := &cli.StringSliceFlag{
		Name:  "buildflags",
		Usage: "build flags to pass to the go compiler. Most commonly use for build flags ie. 'mod upgrade -buildflags=-tags=dev'",
	}
	app := &cli.App{
		Name:  "mod",
		Usage: "upgrade/downgrade semantic import versioning",
		Flags: []cli.Flag{buildFlags},
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
					},
					buildFlags,
				},
			},
			{
				Name:        "downgrade",
				Usage:       "mod downgrade",
				Description: "downgrade go.mod and imports one major version",
				Action:      withExit(downgrade),
				Flags:       []cli.Flag{buildFlags},
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
	return major.Run(".", "upgrade", c.String("mod-name"), c.Int("tag"), c.StringSlice("buildflags"))
}

func downgrade(c *cli.Context) error {
	return major.Run(".", "downgrade", c.String("mod-name"), 0, c.StringSlice("buildflags"))
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

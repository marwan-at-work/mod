package main

import (
	"os"

	"github.com/marwan-at-work/mod/migrate"

	"github.com/marwan-at-work/mod/major"
	"gopkg.in/urfave/cli.v2"
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
				Action:      upgrade,
				Flags: []cli.Flag{
					&cli.IntFlag{
						Name:    "tag",
						Aliases: []string{"t"},
						Value:   0,
					},
				},
			},
			{
				Name:        "downgrade",
				Usage:       "mod downgrade",
				Description: "downgrade go.mod and imports one major version",
				Action:      downgrade,
			},
			{
				Name:        "community-migrate",
				Usage:       "mod community-migrate -token=<github-token>",
				Description: "migrate your +incompatiable dependencies to Go Modules",
				Action:      migrateDeps,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name: "token",
					},
					&cli.IntFlag{
						Name:    "limit",
						Aliases: []string{"l"},
						Value:   -1,
					},
				},
			},
		},
	}

	app.Run(os.Args)
}

func upgrade(c *cli.Context) error {
	return major.Run(".", "upgrade", c.Int("tag"))
}

func downgrade(c *cli.Context) error {
	return major.Run(".", "downgrade", 0)
}

func migrateDeps(c *cli.Context) error {
	return migrate.Run(c.String("token"), c.Int("limit"))
}

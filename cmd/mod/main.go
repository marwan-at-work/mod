package main

import (
	"os"

	"github.com/marwan-at-work/mod/fork"
	"github.com/marwan-at-work/mod/major"
	"github.com/marwan-at-work/mod/migrate"

	cli "gopkg.in/urfave/cli.v2"
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
				},
			},
			{
				Name:        "downgrade",
				Usage:       "mod downgrade",
				Description: "downgrade go.mod and imports one major version",
				Action:      withExit(downgrade),
			},
			{
				Name:        "migrate-deps",
				Usage:       "mod community-migrate -token=<github-token>",
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
			{
				Name:        "fork",
				Usage:       "mod fork <command> <args>",
				Description: "commands for migrating forked libraries",
				Subcommands: []*cli.Command{
					{
						Name:        "rewrite",
						Usage:       "mod fork rewrite <src-repo-import-path-to-overwrite>",
						Description: "rewrite import paths in all .go source files to point to a forked repository import path",
						Action:      withExit(rewriteFork),
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:    "root",
								Aliases: []string{"r"},
								Value:   "./",
							},
						},
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
	return migrate.Run(c.String("token"), c.Int("limit"), c.Bool("test"))
}

func rewriteFork(c *cli.Context) error {
	return fork.Run(c.String("root"), c.Args().First())
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

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/codegangsta/cli"
)

// flags shared between the server and worker subcommands.
var sharedFlags = []cli.Flag{
	cli.StringFlag{
		Name:   "queue",
		Value:  "memory://",
		Usage:  "Build queue to use. Defaults to an in memory queue.",
		EnvVar: "QUEUE",
	},
	cli.StringFlag{
		Name:   "url",
		Value:  "",
		Usage:  "Canonical URL for this instance. Used when adding webhooks to repos.",
		EnvVar: "BASE_URL",
	},
	cli.StringFlag{
		Name:   "logger",
		Value:  "stdout://",
		Usage:  "The logger to use. Available options are `stdout://`, `s3://bucket` or `cloudwatch://`.",
		EnvVar: "LOGGER",
	},
	cli.StringFlag{
		Name:   "db",
		Value:  "",
		Usage:  "A postgres connection string for the database.",
		EnvVar: "DATABASE_URL",
	},
}

func main() {
	app := cli.NewApp()
	app.Name = "conveyor"
	app.Usage = "Build docker images from GitHub repositories"
	app.Action = mainAction
	app.Flags = append(
		sharedFlags,
		append(workerFlags, serverFlags...)...,
	)
	app.Commands = []cli.Command{
		cmdServer,
		cmdWorker,
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func mainAction(c *cli.Context) {
	cy := newConveyor(c)

	worker := make(chan error)
	server := make(chan error)

	go func() {
		worker <- runWorker(cy, c)
	}()

	go func() {
		server <- runServer(cy, c)
	}()

	<-worker
	<-server
}

func must(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

var info = fmt.Printf

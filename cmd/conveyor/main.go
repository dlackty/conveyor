package main

import (
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/codegangsta/cli"
	"github.com/ejholmes/hookshot"
	"github.com/remind101/conveyor"
	"github.com/rlmcpherson/s3gof3r"
)

func main() {
	app := cli.NewApp()
	app.Name = "conveyor"
	app.Usage = "Build docker images from GitHub repositories"
	app.Commands = []cli.Command{
		cmdServer,
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func newBuilder(c *cli.Context) (conveyor.Builder, error) {
	b, err := conveyor.NewDockerBuilderFromEnv()
	if err != nil {
		return nil, err
	}
	b.DryRun = c.Bool("dry")
	b.Image = c.String("builder.image")

	g := conveyor.NewGitHubClient(c.String("github.token"))
	return conveyor.UpdateGitHubCommitStatus(b, g), nil
}

func newServer(c *cli.Context, b conveyor.Builder) (http.Handler, error) {
	b = conveyor.BuildAsync(b)
	s := conveyor.NewServer(b)

	f, err := logFactory(c.String("logger"))
	if err != nil {
		return nil, err
	}
	s.LogFactory = f

	return hookshot.Authorize(s, c.String("github.secret")), nil
}

func logFactory(uri string) (f conveyor.LogFactory, err error) {
	var u *url.URL
	u, err = url.Parse(uri)
	if err != nil {
		return nil, err
	}

	switch u.Scheme {
	case "s3":
		f, err = conveyor.S3Logger(u.Host, s3gof3r.EnvKeys)
		if err != nil {
			return
		}
	}

	// f = conveyor.MultiLogger(conveyor.StdoutLogger, f)
	return
}

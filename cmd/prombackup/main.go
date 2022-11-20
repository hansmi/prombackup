package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"os"

	"github.com/google/subcommands"
	"github.com/hansmi/prombackup/api"
	"github.com/hansmi/prombackup/client"
	"github.com/hansmi/prombackup/internal/clientcli"
	"github.com/hansmi/prombackup/internal/clientcli/create"
	"github.com/hansmi/prombackup/internal/clientcli/prune"
)

func main() {
	showVersion := flag.Bool("version", false, "Output version information and exit.")
	serverURL := flag.String("server", os.Getenv("PROMBACKUP_ENDPOINT"),
		"Server endpoint URL. Defaults to the PROMBACKUP_ENDPOINT environment variable.")

	subcommands.Register(subcommands.HelpCommand(), "")
	subcommands.Register(subcommands.FlagsCommand(), "")
	subcommands.Register(subcommands.CommandsCommand(), "")
	subcommands.Register(&create.Command{}, "")
	subcommands.Register(&prune.Command{}, "")

	flag.Parse()

	if *showVersion {
		printVersion(os.Stdout)
		return
	}

	rt := &clientcli.Runtime{
		NewClient: func() (api.Interface, error) {
			if *serverURL == "" {
				return nil, errors.New("missing top-level -server flag")
			}

			return client.New(client.Options{
				Address:   *serverURL,
				Logger:    log.Default(),
				UserAgent: clientUserAgent(),
			})
		},
	}

	os.Exit(int(subcommands.Execute(context.Background(), rt)))
}

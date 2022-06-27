package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ONSdigital/dp-geodata-api/cmd/filescli/app"
	"github.com/ONSdigital/dp-geodata-api/cmd/filescli/cli"
	"github.com/ONSdigital/dp-geodata-api/cmd/filescli/config"

	"github.com/ONSdigital/dp-api-clients-go/v2/files"
	"github.com/ONSdigital/dp-api-clients-go/v2/upload"
)

func main() {
	help := flag.Bool("h", false, "help")
	cfgname := flag.String("C", "config.yml", "host config file")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [command-options] upload|setid|getstate|publish|getpub|getweb [subcommand-options]\n", filepath.Base(os.Args[0]))
		flag.PrintDefaults()
	}
	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(0)
	}
	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(2)
	}

	cfg, err := config.Load(*cfgname)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}

	ctx := context.Background()

	identToken, err := app.Identify(
		ctx,
		os.Getenv("IDENTITY_TOKEN"),
		cfg.Hosts.Identity,
		os.Getenv("IDENTITY_EMAIL"),
		os.Getenv("IDENTITY_PASSWORD"),
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		if !errors.Is(err, app.ErrIdentifyWarning) {
			os.Exit(1)
		}
	}

	loginToken, err := app.Login(
		ctx,
		os.Getenv("FLORENCE_TOKEN"),
		cfg.Hosts.Zebedee,
		os.Getenv("FLORENCE_EMAIL"),
		os.Getenv("FLORENCE_PASSWORD"),
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		if !errors.Is(err, app.ErrLoginWarning) {
			os.Exit(1)
		}
	}

	cli := &cli.CLI{
		App: &app.App{
			IdentToken:     identToken,
			LoginToken:     loginToken,
			FilesURL:       cfg.Hosts.Files,
			UploadURL:      cfg.Hosts.Upload,
			WebDownloadURL: cfg.Hosts.DownloadWeb,
			PubDownloadURL: cfg.Hosts.DownloadPublishing,
			UploadClient:   upload.NewAPIClient(cfg.Hosts.Upload, identToken),
			FilesClient:    files.NewAPIClient(cfg.Hosts.Files, identToken),
			Out:            os.Stdout,
		},
	}

	argv := flag.Args()[1:]
	switch flag.Arg(0) {
	case "upload":
		err = cli.Upload(ctx, argv)
	case "setid":
		err = cli.SetID(ctx, argv)
	case "getstate":
		err = cli.GetState(ctx, argv)
	case "publish":
		err = cli.Publish(ctx, argv)
	case "getpub":
		err = cli.GetPub(ctx, argv)
	case "getweb":
		err = cli.GetWeb(ctx, argv)
	default:
		flag.Usage()
		os.Exit(2)
	}

	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

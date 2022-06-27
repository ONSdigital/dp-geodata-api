package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/ONSdigital/dp-geodata-api/cmd/filescli/app"
)

type CLI struct {
	App *app.App
}

func (c *CLI) Upload(ctx context.Context, argv []string) error {
	flagset := flag.NewFlagSet("upload", flag.ExitOnError)
	help := flagset.Bool("h", false, "help")
	collectionId := flagset.String("c", "", "collection id (optional)")
	mimetype := flagset.String("m", "", "mime type (optional; detected if not given)")
	title := flagset.String("t", "", "title")
	ispublishable := flagset.Bool("p", false, "is publishable")
	license := flagset.String("l", "MIT", "license type")
	licenseurl := flagset.String("L", "https://opensource.org/licenses/MIT", "license URL")
	flagset.Usage = func() {
		fmt.Fprintln(flag.CommandLine.Output(), "Usage: upload [options] <local-path> <remote-path>")
		flagset.PrintDefaults()
	}
	flagset.Parse(argv)

	if *help {
		flagset.Usage()
		return nil
	}
	if flagset.NArg() < 2 {
		return errors.New("upload: local and remote paths required")
	}
	localpath := flagset.Arg(0)
	remotepath := flagset.Arg(1)
	if *title == "" || *license == "" || *licenseurl == "" || localpath == "" || remotepath == "" {
		return errors.New("upload: missing required argument")
	}
	if *collectionId == "" {
		collectionId = nil
	}

	if *mimetype == "" {
		mime, err := detectContentType(localpath)
		if err != nil {
			return fmt.Errorf("%s: %w", localpath, err)
		}
		*mimetype = mime
	}

	f, err := os.Open(localpath)
	if err != nil {
		return err
	}
	rc := &multiCloser{RC: f}
	defer rc.Close()

	info, err := f.Stat()
	if err != nil {
		return err
	}

	return c.App.Upload(ctx, collectionId, *mimetype, *title, *license, *licenseurl, *ispublishable, rc, info.Size(), remotepath)
}

func (c *CLI) SetID(ctx context.Context, argv []string) error {
	flagset := flag.NewFlagSet("setid", flag.ExitOnError)
	help := flagset.Bool("h", false, "help")
	id := flagset.String("c", "", "new collection id")
	flagset.Usage = func() {
		fmt.Fprintln(flag.CommandLine.Output(), "Usage: setid -c <collection-id> <path>")
		flagset.PrintDefaults()
	}
	flagset.Parse(argv)
	if *help {
		flagset.Usage()
		return nil
	}
	if flagset.NArg() == 0 {
		return errors.New("setid: path required")
	}
	if *id == "" {
		return errors.New("setid: new collection id required")
	}

	return c.App.SetID(ctx, flagset.Arg(0), *id)
}

func (c *CLI) GetState(ctx context.Context, argv []string) error {
	flagset := flag.NewFlagSet("getstate", flag.ExitOnError)
	help := flagset.Bool("h", false, "help")
	flagset.Usage = func() {
		fmt.Fprintln(flag.CommandLine.Output(), "Usage: getstate <path>")
		flagset.PrintDefaults()
	}
	flagset.Parse(argv)
	if *help {
		flagset.Usage()
		return nil
	}
	if flagset.NArg() == 0 {
		return errors.New("path required")
	}

	return c.App.GetState(ctx, flagset.Arg(0))
}

func (c *CLI) Publish(ctx context.Context, argv []string) error {
	flagset := flag.NewFlagSet("setid", flag.ExitOnError)
	help := flagset.Bool("h", false, "help")
	id := flagset.String("c", "", "collection id")
	flagset.Usage = func() {
		fmt.Fprintln(flag.CommandLine.Output(), "Usage: publish -c <collection-id>")
		flagset.PrintDefaults()
	}
	flagset.Parse(argv)
	if *help {
		flagset.Usage()
		return nil
	}

	return c.App.Publish(ctx, *id)
}

func (c *CLI) GetPub(ctx context.Context, argv []string) error {
	flagset := flag.NewFlagSet("getpub", flag.ExitOnError)
	help := flagset.Bool("h", false, "help")
	flagset.Usage = func() {
		fmt.Fprintln(flag.CommandLine.Output(), "Usage: getpub <path>")
		flagset.PrintDefaults()
	}
	flagset.Parse(argv)
	if *help {
		flagset.Usage()
		return nil
	}
	if flagset.NArg() == 0 {
		flagset.Usage()
		return errors.New("path required")
	}

	return c.App.GetPub(ctx, flagset.Arg(0))
}

func (c *CLI) GetWeb(ctx context.Context, argv []string) error {
	flagset := flag.NewFlagSet("getweb", flag.ExitOnError)
	help := flagset.Bool("h", false, "help")
	flagset.Usage = func() {
		fmt.Fprintln(flag.CommandLine.Output(), "Usage: getweb <path>")
		flagset.PrintDefaults()
	}
	flagset.Parse(argv)
	if *help {
		flagset.Usage()
		return nil
	}
	if flagset.NArg() == 0 {
		flagset.Usage()
		return errors.New("path required")
	}

	return c.App.GetWeb(ctx, flagset.Arg(0))
}

func detectContentType(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	buf := make([]byte, 512)
	_, err = f.Read(buf)
	if err != nil {
		return "", err
	}

	return http.DetectContentType(buf), nil
}

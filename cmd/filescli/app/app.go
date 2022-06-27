// app implements static file operations such as upload, publish, and download.
//
// Some operations use the dp-api-clients-go package, and some use regular http calls.
// This package presents a more uniform set of methods for the operations
// required by the cli.

package app

import (
	"context"
	"errors"
	"io"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/ONSdigital/dp-api-clients-go/v2/upload"
)

//go:generate go run github.com/matryer/moq@v0.2.7 -out mock/filer.go -pkg mock . Filer
// A Filer knows how to set collection ids and publish (eg files.Client)
type Filer interface {
	SetCollectionID(ctx context.Context, path, id string) error
	PublishCollection(ctx context.Context, id string) error
}

//go:generate go run github.com/matryer/moq@v0.2.7 -out mock/uploader.go -pkg mock . Uploader
// An Uploader knows how to upload a file in chunks (eg upload.Clients)
type Uploader interface {
	Upload(ctx context.Context, f io.ReadCloser, meta upload.Metadata) error
}

type App struct {
	IdentToken     string    // optional identity token (see Identify)
	LoginToken     string    // optional login token (Login)
	FilesURL       string    // URL of Files Service
	UploadURL      string    // URL of Upload Service
	WebDownloadURL string    // URL of Web Download Service
	PubDownloadURL string    // URL of Publishing Download Seervice
	UploadClient   Uploader  // *upload.Client
	FilesClient    Filer     // *files.Client
	Out            io.Writer // destination for file downloads
}

// splitPath splits a remote pathname into directory and filename parts.
//
// Direct http requests expect a single remote path, but some library
// functions expect a separate path and filename.
// So splitPath lets our functions always accept a single "remote" path,
// regardless of how the underlying function wants it.
func splitPath(path string) (string, string, error) {
	if filepath.IsAbs(path) {
		return "", "", errors.New("path must not be absolute")
	}

	dir, name := filepath.Split(path)
	dir = strings.TrimRight(dir, "/")
	if name == "" {
		return "", "", errors.New("path must include filename")
	}
	return dir, name, nil
}

// isRegular is true if info describes a regular file (ie not a directory, etc)
func isRegular(info fs.FileInfo) bool {
	filetype := info.Mode() & fs.ModeType
	return filetype == 0
}

package app

import (
	"context"
	"io"

	"github.com/ONSdigital/dp-api-clients-go/v2/upload"
)

// Upload uploads a file using the Upload Seervice.
//
// id is either a collection id, or nil.
// When id is nil, a collection id may be assigned to the file later.
//
// mime is the file's mime type.
//
// title is an arbitrary title for the file.
//
// lic and licurl are license and license url.
// For example, "BSD" and "https://opensource.org/licenses/BSD-3-Clause"
//
// The file to be uploaded will be read from rc.
// (Although rc is an io.ReadCloser, I don't think it is actually closed.)
//
// size is the size of the file to be uploaded.
//
// remote is the pathname on the destination.
func (a *App) Upload(ctx context.Context, id *string, mime, title, lic, licurl string, ispublishable bool, rc io.ReadCloser, size int64, remote string) error {
	dir, name, err := splitPath(remote)
	if err != nil {
		return err
	}

	meta := upload.Metadata{
		CollectionID:  id,
		FileName:      name,
		Path:          dir,
		IsPublishable: ispublishable,
		Title:         title,
		FileSizeBytes: size,
		FileType:      mime,
		License:       lic,
		LicenseURL:    licurl,
	}

	return a.UploadClient.Upload(ctx, rc, meta)
}

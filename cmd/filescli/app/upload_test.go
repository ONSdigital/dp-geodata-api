package app_test

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/ONSdigital/dp-geodata-api/cmd/filescli/app"
	"github.com/ONSdigital/dp-geodata-api/cmd/filescli/app/mock"

	"github.com/ONSdigital/dp-api-clients-go/v2/upload"
	. "github.com/smartystreets/goconvey/convey"
)

func Test_Upload(t *testing.T) {
	Convey("A bad remote name should return an error", t, func() {
		a := app.App{}
		err := a.Upload(context.Background(), nil, "", "", "", "", false, nil, 0, "/path")
		So(err, ShouldBeError, "path must not be absolute")
	})

	Convey("An upload error should be propagated up", t, func() {
		wanterr := errors.New("error from underlying Upload")
		mockedUploader := &mock.UploaderMock{
			UploadFunc: func(ctx context.Context, f io.ReadCloser, meta upload.Metadata) error {
				return wanterr
			},
		}
		a := app.App{
			UploadClient: mockedUploader,
		}
		err := a.Upload(context.Background(), nil, "", "", "", "", false, nil, 0, "dir/file")
		So(err, ShouldBeError, wanterr)
	})

	Convey("Metadata fields should be sent to underlying Upload", t, func() {
		rc := os.Stdin
		id := "id"
		meta := upload.Metadata{
			CollectionID:  &id,
			FileName:      "name",
			Path:          "path",
			IsPublishable: true,
			Title:         "title",
			FileSizeBytes: 100,
			FileType:      "text/plain",
			License:       "BSD",
			LicenseURL:    "https://opensource.org/licenses/BSD-3-Clause",
		}
		ctx := context.Background()
		mockedUploader := &mock.UploaderMock{
			UploadFunc: func(myctx context.Context, myrc io.ReadCloser, mymeta upload.Metadata) error {
				if myctx != ctx {
					return errors.New("correct context not passed")
				}
				if myrc != rc {
					return errors.New("correct ReadCloser not passed")
				}
				if mymeta != meta {
					return errors.New("correct meta not passed")
				}
				return nil
			},
		}
		a := app.App{
			UploadClient: mockedUploader,
		}
		err := a.Upload(
			ctx,
			meta.CollectionID,
			meta.FileType,
			meta.Title,
			meta.License,
			meta.LicenseURL,
			meta.IsPublishable,
			rc,
			meta.FileSizeBytes,
			filepath.Join(meta.Path, meta.FileName),
		)
		So(err, ShouldBeNil)
	})
}

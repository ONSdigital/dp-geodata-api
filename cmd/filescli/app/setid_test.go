package app_test

import (
	"context"
	"errors"
	"testing"

	"github.com/ONSdigital/dp-geodata-api/cmd/filescli/app"
	"github.com/ONSdigital/dp-geodata-api/cmd/filescli/app/mock"

	. "github.com/smartystreets/goconvey/convey"
)

func Test_SetID(t *testing.T) {
	Convey("An absolute remote name should return an error", t, func() {
		a := app.App{}
		err := a.SetID(context.Background(), "/path", "")
		So(err, ShouldBeError, "path must not be absolute")
	})

	Convey("Correct parameters should be sent to underlying SetCollectionID", t, func() {
		const (
			id   = "id"
			path = "dir/name"
		)
		ctx := context.Background()
		mockedFiler := &mock.FilerMock{
			SetCollectionIDFunc: func(myctx context.Context, mypath, myid string) error {
				if myctx != ctx {
					return errors.New("correct context not passed")
				}
				if mypath != path {
					return errors.New("correct path not passed")
				}
				if myid != id {
					return errors.New("correct id not passed")
				}
				return nil
			},
		}
		a := app.App{
			FilesClient: mockedFiler,
		}
		err := a.SetID(ctx, path, id)
		So(err, ShouldBeNil)
	})
}

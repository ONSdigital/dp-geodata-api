package app_test

import (
	"context"
	"errors"
	"testing"

	"github.com/ONSdigital/dp-geodata-api/cmd/filescli/app"
	"github.com/ONSdigital/dp-geodata-api/cmd/filescli/app/mock"

	. "github.com/smartystreets/goconvey/convey"
)

func Test_Publish(t *testing.T) {
	Convey("Correct parameters should be sent to underlying PublishCollection", t, func() {
		const id = "id"
		ctx := context.Background()
		mockedFiler := &mock.FilerMock{
			PublishCollectionFunc: func(myctx context.Context, myid string) error {
				if myctx != ctx {
					return errors.New("context not passed correctly")
				}
				if myid != id {
					return errors.New("id not passed correctly")
				}
				return nil
			},
		}
		a := app.App{
			FilesClient: mockedFiler,
		}
		err := a.Publish(ctx, id)
		So(err, ShouldBeNil)
	})
}

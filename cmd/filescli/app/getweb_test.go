package app_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ONSdigital/dp-geodata-api/cmd/filescli/app"

	. "github.com/smartystreets/goconvey/convey"
)

func Test_GetWeb(t *testing.T) {

	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "expected Get", http.StatusBadRequest)
			return
		}
		switch r.URL.Path {
		case "/downloads-new/not/found":
			http.NotFound(w, r)
		case "/downloads-new/return/error":
			http.Error(w, "got an error", http.StatusInternalServerError)
		case "/downloads-new/dir/file":
			fmt.Fprintf(w, "contents of dir/file")
		}
	}

	ctx := context.Background()

	Convey("Tests without a server", t, func() {
		a := app.App{}
		Convey("An absolute remote name should return an error", func() {
			err := a.GetWeb(ctx, "/path")
			So(err, ShouldBeError, "path must not be absolute")
		})

		Convey("A missing WebDownloadURL should return an error", func() {
			err := a.GetWeb(ctx, "any/thing")
			So(err, ShouldNotBeNil) // url.Error
		})
	})

	Convey("Tests against a server", t, func() {
		ts := httptest.NewServer(http.HandlerFunc(handler))
		defer ts.Close()

		Convey("Errors", func() {
			a := app.App{
				WebDownloadURL: ts.URL,
			}

			Convey("A remote 404 should return a not found error", func() {
				err := a.GetWeb(ctx, "not/found")
				So(err, ShouldBeError, "file not found on remote")
			})

			Convey("An http error should return an error message", func() {
				err := a.GetWeb(ctx, "return/error")
				So(err, ShouldBeError, "got an error\n")
			})
		})

		Convey("Successful request", func() {
			var buf strings.Builder
			a := app.App{
				WebDownloadURL: ts.URL,
				Out:            &buf,
			}
			err := a.GetWeb(ctx, "dir/file")
			So(err, ShouldBeNil)
			So(buf.String(), ShouldEqual, "contents of dir/file")
		})
	})
}

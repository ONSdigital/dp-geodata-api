package app

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func Test_GetState(t *testing.T) {
	const identToken = "expected-identify-token"
	expResp := stateResponse{
		Path:          "dir/file",
		IsPublishable: true,
		Title:         "title",
		Size:          100,
		MimeType:      "text/plain",
		License:       "MIT",
		LicenseURL:    "https://opensource.org/licenses/MIT",
		State:         "PUBLISHED",
		Etag:          "whatever",
	}
	errResp := stateResponse{
		Errors: []rerror{
			{
				Code:        "error code",
				Description: "description of the error",
			},
		},
	}

	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "expected Get", http.StatusBadRequest)
			return
		}
		if r.Header.Get("Authorization") != "Bearer "+identToken {
			http.Error(w, "", http.StatusUnauthorized)
			return
		}

		returnjson := func(t *testing.T, w http.ResponseWriter, resp *stateResponse) {
			w.Header().Add("Content-Type", "application/json")
			enc := json.NewEncoder(w)
			if err := enc.Encode(resp); err != nil {
				t.Fatal(err)
			}
		}

		switch r.URL.Path {
		case "/files/return/error":
			http.Error(w, "internal error", http.StatusInternalServerError)
		case "/files/not/json":
			w.Write([]byte("this is not json as expected"))
		case "/files/errors/in/json":
			returnjson(t, w, &errResp)
		case "/files/dir/file":
			returnjson(t, w, &expResp)
		default:
			http.NotFound(w, r)
		}
	}

	ctx := context.Background()

	Convey("Tests without a server", t, func() {
		Convey("An absolute remote name should return an error", func() {
			a := App{}
			err := a.GetState(ctx, "/path")
			So(err, ShouldBeError, "path must not be absolute")
		})

		Convey("Wrong Files service URL should return an error", func() {
			a := App{
				IdentToken: identToken,
				FilesURL:   "http://localhost:0",
			}
			err := a.GetState(ctx, "dir/file")
			So(err, ShouldNotBeNil) // url.Error
		})
	})

	Convey("Tests against a server", t, func() {
		ts := httptest.NewServer(http.HandlerFunc(handler))
		defer ts.Close()

		Convey("Errors", func() {
			a := App{
				IdentToken: identToken,
				FilesURL:   ts.URL,
			}

			Convey("404 should return not found error", func() {
				err := a.GetState(ctx, "not/found")
				So(err, ShouldBeError, "file not found on remote")
			})

			Convey("non-200 should return error message", func() {
				err := a.GetState(ctx, "return/error")
				So(err, ShouldBeError, "internal error\n")
			})

			Convey("non-json response should return an error", func() {
				err := a.GetState(ctx, "not/json")
				So(err, ShouldBeError, "could not parse response from remote")
			})

			Convey("error messeges in json should return an error", func() {
				err := a.GetState(ctx, "errors/in/json")
				So(err, ShouldBeError, "description of the error")
			})
		})

		Convey("Successful request", func() {
			var buf bytes.Buffer
			a := App{
				IdentToken: identToken,
				FilesURL:   ts.URL,
				Out:        &buf,
			}
			err := a.GetState(ctx, "dir/file")
			So(err, ShouldBeNil)
			var resp stateResponse
			err = json.Unmarshal(buf.Bytes(), &resp)
			So(err, ShouldBeNil)
			So(resp, ShouldResemble, expResp)
		})
	})
}

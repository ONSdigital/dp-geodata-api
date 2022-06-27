package app

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func Test_Login(t *testing.T) {
	const (
		loginToken = "expected-login-token"
		email      = "somebody@example.com"
		password   = "correct-password"
	)

	ctx := context.Background()

	Convey("Tests without a server", t, func() {
		Convey("returns original token if non-blank", func() {
			token, err := Login(ctx, loginToken, "", "", "")
			So(err, ShouldBeNil)
			So(token, ShouldEqual, loginToken)
		})

		Convey("returns warning when there are no auth variables", func() {
			_, err := Login(ctx, "", "", "", "")
			So(err, ShouldWrap, ErrLoginWarning)
		})

		Convey("returns error with wrong Identity Service URL", func() {
			_, err := Login(ctx, "", "http://localhost;0", email, password)
			So(err, ShouldNotBeNil) // url.Error
		})
	})

	Convey("Tests against a server", t, func() {
		handler := func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				http.Error(w, "expected Post", http.StatusBadRequest)
				return
			}
			if r.URL.Path != "/login" {
				http.Error(w, "expected /login path", http.StatusBadRequest)
				return
			}
			if r.Header.Get("Content-Type") != "application/json" {
				http.Error(w, "expected Content-Type application/json", http.StatusBadRequest)
				return
			}
			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, string(body), http.StatusInternalServerError)
				return
			}
			var creds loginCreds
			if err = json.Unmarshal(body, &creds); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if creds.Email != email || creds.Password != password {
				// XXX verify this is what real server does
				http.Error(w, "bad credentials", http.StatusUnauthorized)
				return
			}
			w.Write([]byte(`"` + loginToken + `"`))
		}

		ts := httptest.NewServer(http.HandlerFunc(handler))
		defer ts.Close()

		Convey("non-200 returns an error", func() {
			_, err := Login(ctx, "", ts.URL, email, "")
			So(err, ShouldBeError, "bad credentials\n")
		})

		Convey("correct token returned", func() {
			token, err := Login(ctx, "", ts.URL, email, password)
			So(err, ShouldBeNil)
			So(token, ShouldEqual, loginToken)
		})
	})
}

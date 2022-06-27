package app

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func Test_Identify(t *testing.T) {
	const (
		identToken = "expected-identify-token"
		email      = "somebody@example.com"
		password   = "correct-password"
		omitHeader = "omit-header"  // special "password" to trigger an error in http handler
		noBearer   = "wrong-scheme" // "
		send200    = "non-200"      // "
	)

	handler := func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "expected Post", http.StatusBadRequest)
			return
		}
		if r.URL.Path != "/v1/tokens" {
			http.Error(w, "expected /v1/token path", http.StatusBadRequest)
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
		var creds identifyCreds
		if err = json.Unmarshal(body, &creds); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if creds.Email != email {
			// XXX verify this is what the real server does
			http.Error(w, "bad credentials", http.StatusUnauthorized)
			return
		}
		// overload the password field to cause certain errors
		switch creds.Password {
		case password:
			w.Header().Set("Authorization", "Bearer expected-identify-token")
			w.WriteHeader(http.StatusCreated)
		case omitHeader:
			w.WriteHeader(http.StatusCreated)
		case noBearer:
			w.Header().Set("Authorization", "Notbearer placeholder")
			w.WriteHeader(http.StatusCreated)
		case send200:
			// send correct Authorization header, but 200 status instead of expected 201
			w.Header().Set("Authorization", "Bearer expected-identify-token")
			fmt.Fprintf(w, "status 200")
		default:
			http.Error(w, "bad credentials", http.StatusUnauthorized)
		}
	}

	ctx := context.Background()

	Convey("Tests without a server", t, func() {
		Convey("Missing auth variables should return a warning", func() {
			_, err := Identify(ctx, "", "", "", "")
			So(err, ShouldWrap, ErrIdentifyWarning)
		})

		Convey("Wrong URL should return an error", func() {
			_, err := Identify(ctx, "", "http://localhost:0", email, password)
			So(err, ShouldNotBeNil) // url.Error
		})

		Convey("Existing token should be returned", func() {
			token, err := Identify(ctx, identToken, "", "", "")
			So(err, ShouldBeNil)
			So(token, ShouldEqual, identToken)
		})
	})

	Convey("Tests against a server", t, func() {
		ts := httptest.NewServer(http.HandlerFunc(handler))
		defer ts.Close()

		Convey("Non-201 status should return an error", func() {
			_, err := Identify(ctx, "", ts.URL, email, send200)
			So(err, ShouldBeError, "status 200")
		})

		Convey("Missing Authorization header should return an error", func() {
			_, err := Identify(ctx, "", ts.URL, email, omitHeader)
			So(err, ShouldBeError, "missing or bad Authorization header returned from identity service")
		})

		Convey("Non-Bearer scheme in Authorization header should return an error", func() {
			_, err := Identify(ctx, "", ts.URL, email, noBearer)
			So(err, ShouldBeError, "expected Bearer scheme in Authorization header returned from identity service")
		})

		Convey("Correct credentials return token", func() {
			token, err := Identify(ctx, "", ts.URL, email, password)
			So(err, ShouldBeNil)
			So(token, ShouldEqual, identToken)
		})
	})
}

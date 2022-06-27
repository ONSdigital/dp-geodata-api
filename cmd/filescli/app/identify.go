package app

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/ONSdigital/dp-geodata-api/sentinel"
)

const ErrIdentifyWarning = sentinel.Sentinel("could not get authentication token")

type identifyCreds struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Identify attempts to authorise the user and get an identity token.
// If token is non-empty, it is returned without making any server requests.
// Otherwise email and password are passed to server and the results are returned.
// If there is not enough information to request a token from the server,
// ErrIdentifyWarning is returned.
// (This is just a warning because certain operations don't need a token.)
func Identify(ctx context.Context, token, server, email, password string) (string, error) {
	if token != "" {
		return token, nil // already have an identity token, just return it
	}
	if server == "" || email == "" { // password may be blank
		return "", fmt.Errorf("%w: identity server, email and password required", ErrIdentifyWarning)
	}

	creds := identifyCreds{
		Email:    email,
		Password: password,
	}

	url := server + "/v1/tokens"
	body, err := json.Marshal(&creds)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}
		return "", errors.New(string(body))
	}

	// body contains additional information, but we don't use it;
	// we only use the Authorization header
	token = resp.Header.Get("Authorization")
	scheme, params, found := strings.Cut(token, " ")
	if !found {
		return "", errors.New("missing or bad Authorization header returned from identity service")
	}
	if !strings.EqualFold(scheme, "Bearer") {
		return "", errors.New("expected Bearer scheme in Authorization header returned from identity service")
	}
	return params, nil
}

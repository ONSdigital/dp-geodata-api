package app

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/ONSdigital/dp-geodata-api/sentinel"
)

const ErrLoginWarning = sentinel.Sentinel("could not get login token")

type loginCreds struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Login attempts to authorise the Florence user and get a token.
// If token is non-empty, it is returned without making any server requests.
// Otherwise email and password are passed to server and the results are returned.
// If there is not enough information to request a token from the server,
// ErrLoginWarning is returned.
// (This is just a warning because certain operations don't need a token.)
//
// (Note: "Login" may not be correct term for this operation.)
func Login(ctx context.Context, token, server, email, password string) (string, error) {
	if token != "" {
		return token, nil // already have a login token, just return it
	}
	if server == "" || email == "" {
		return "", fmt.Errorf("%w: login server, email and password required", ErrLoginWarning)
	}

	creds := loginCreds{
		Email:    email,
		Password: password,
	}

	url := server + "/login"
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

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", errors.New(string(body))
	}

	return string(bytes.Trim(body, `"`)), nil
}

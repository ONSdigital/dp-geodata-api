package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type stateResponse struct {
	Errors        []rerror `json:"errors"`
	Path          string   `json:"path"`
	IsPublishable bool     `json:"is_publishable"`
	Title         string   `json:"title"`
	Size          int      `json:"size_in_bytes"`
	MimeType      string   `json:"type"`
	License       string   `json:"license"`
	LicenseURL    string   `json:"license_url"`
	State         string   `json:"state"`
	Etag          string   `json:"etag"`
}
type rerror struct {
	Code        string `json:"code"`
	Description string `json:"description"`
}

// GetState gets the state of the remote file named path.
// a.FilesURL and a.IdentToken are required.
// The server response is written to a.Out.
func (a *App) GetState(ctx context.Context, path string) error {
	dir, name, err := splitPath(path)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/files/%s/%s", a.FilesURL, dir, name)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", "Bearer "+a.IdentToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusNotFound {
		return errors.New("file not found on remote")
	}
	if resp.StatusCode != http.StatusOK {
		return errors.New(string(body))
	}

	var state stateResponse
	err = json.Unmarshal(body, &state)
	if err != nil {
		return errors.New("could not parse response from remote")
	}
	if len(state.Errors) != 0 {
		return errors.New(listErrors(state.Errors))
	}

	_, err = a.Out.Write(body)
	return err
}

func listErrors(rerrs []rerror) string {
	var msgs []string
	for _, rerr := range rerrs {
		msgs = append(msgs, rerr.Description)
	}
	return strings.Join(msgs, ": ")
}

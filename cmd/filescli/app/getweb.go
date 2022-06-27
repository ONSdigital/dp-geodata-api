package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// GetWeb downloads a file from Web.
// path is the remote pathname.
// a.WebDownloadURL is required.
func (a *App) GetWeb(ctx context.Context, path string) error {
	dir, name, err := splitPath(path)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/downloads-new/%s/%s", a.WebDownloadURL, dir, name)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return errors.New("file not found on remote")
	}
	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return errors.New(string(body))
	}

	_, err = io.Copy(a.Out, resp.Body)
	return err
}

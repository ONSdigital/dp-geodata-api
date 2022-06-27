package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// GetPub downloads a file from Publishing.
// path is the remote pathname.
// a.PubDownloadURL, a.IdentToken and a.LoginToken are required.
func (a *App) GetPub(ctx context.Context, path string) error {
	dir, name, err := splitPath(path)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/downloads-new/%s/%s", a.PubDownloadURL, dir, name)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+a.IdentToken)
	req.Header.Set("X-Florence-Token", a.LoginToken)

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

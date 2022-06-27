package app

import (
	"context"
	"path/filepath"
)

// SetID sets the collection id of a remote file.
// a.FilesClient is required.
func (a *App) SetID(ctx context.Context, path, id string) error {
	dir, name, err := splitPath(path)
	if err != nil {
		return err
	}
	return a.FilesClient.SetCollectionID(ctx, filepath.Join(dir, name), id)
}

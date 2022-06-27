package app

import "context"

// Publish marks a collection id as published.
// a.FilesClient required.
func (a *App) Publish(ctx context.Context, id string) error {
	return a.FilesClient.PublishCollection(ctx, id)
}

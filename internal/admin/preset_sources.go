package admin

import (
	"context"

	"github.com/kagi-labs/agentnyk-maisternia/internal/presetsources"
)

func (l Loader) AddPresetSource(location string) (presetsources.Source, error) {
	return (presetsources.Manager{
		Home:   l.Home,
		Getenv: l.Getenv,
	}).Add(context.Background(), presetsources.AddRequest{Location: location})
}

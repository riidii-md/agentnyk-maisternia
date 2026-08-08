package app

import (
	"fmt"
	"os"

	"github.com/kagi-labs/agentnyk-maisternia/internal/repository"
)

func resolveRepositoryOption(value, home string) (string, error) {
	selection, err := resolveRepositorySelection(value, home)
	if err != nil {
		return "", err
	}
	return selection.Path, nil
}

func resolveRepositorySelection(value, home string) (repository.Selection, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return repository.Selection{}, fmt.Errorf("resolve current directory: %w", err)
	}
	selection, err := repository.Resolve(repository.Options{
		Explicit: value,
		Home:     home,
		Cwd:      cwd,
	})
	if err != nil {
		return repository.Selection{}, err
	}
	return selection, nil
}

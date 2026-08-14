package maisternia

import (
	"embed"
	"io/fs"
)

// embeddedCatalog contains the versioned provider, preset, collection, policy,
// and workflow definitions shipped with every maisternia binary.
//
//go:embed config
var embeddedCatalog embed.FS

// CatalogFS returns the read-only configuration catalog embedded in the binary.
func CatalogFS() fs.FS {
	return embeddedCatalog
}

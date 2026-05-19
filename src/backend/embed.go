// Package aiengine provides the embedded frontend assets.
package aiengine

import "embed"

//go:embed frontend-dist
var Files embed.FS

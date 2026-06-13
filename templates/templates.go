// Package templates exposes the bundled HTML report templates through an
// embedded filesystem so the binary can always render reports even when no
// template directory is present on disk (for example in a custom image where
// /app/templates is missing). The on-disk --template-dir override still takes
// precedence when the requested file exists there.
package templates

import "embed"

//go:embed *.html
var FS embed.FS

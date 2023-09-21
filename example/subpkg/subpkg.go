//go:build something

package subpkg

import (
	// running mod upgrade|downgrade from the
	// example folder, will update this path
	// according to semantic import versioning.
	_ "github.com/marwan-at-work/mod/example"
)

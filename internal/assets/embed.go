// Package assets provides embedded application resources.
package assets

import (
	_ "embed"
)

//go:embed icons/guanaco-logo.svg
var LogoSVG []byte

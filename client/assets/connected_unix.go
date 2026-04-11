//go:build !windows

package assets

import _ "embed"

//go:embed connected.png
var Connected []byte

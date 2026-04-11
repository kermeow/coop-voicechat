//go:build !windows

package assets

import _ "embed"

//go:embed disconnected.png
var Disconnected []byte

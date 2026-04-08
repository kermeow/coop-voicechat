//go:build linux || darwin

package assets

import _ "embed"

//go:embed microphone.png
var Microphone []byte

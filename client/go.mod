module coop-voicechat

go 1.26.1

require (
	github.com/BurntSushi/toml v1.6.0
	github.com/energye/systray v1.0.3
	github.com/gordonklaus/portaudio v0.0.0-20260203164431-765aa7dfa631
	github.com/hraban/opus v0.0.0-20251117090126-c76ea7e21bf3
	github.com/postfinance/single v0.0.2
	github.com/quartercastle/vector v0.2.0
	github.com/gopxl/beep/v2 v2.1.1
)

require (
	github.com/godbus/dbus/v5 v5.2.2 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	golang.org/x/sys v0.43.0 // indirect
)

replace github.com/gordonklaus/portaudio => github.com/kermeow/static-portaudio v0.0.0

replace github.com/hraban/opus => github.com/kermeow/static-opus v0.0.3

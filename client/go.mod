module coop-voicechat

go 1.26.2

require (
	github.com/energye/systray v1.0.3
	github.com/gopxl/beep/v2 v2.1.1
	github.com/gordonklaus/portaudio v0.0.0-20260203164431-765aa7dfa631
	github.com/hraban/opus v0.0.0-20251117090126-c76ea7e21bf3
	github.com/kermeow/rnnoise v0.0.2
	github.com/postfinance/single v0.0.2
	github.com/quartercastle/vector v0.2.0
)

require (
	github.com/ebitengine/oto/v3 v3.3.2 // indirect
	github.com/ebitengine/purego v0.8.0 // indirect
	github.com/godbus/dbus/v5 v5.1.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	golang.org/x/sys v0.25.0 // indirect
)

replace github.com/energye/systray => github.com/kermeow/systray v0.0.0

replace github.com/gordonklaus/portaudio => github.com/kermeow/static-portaudio v0.0.2

replace github.com/hraban/opus => github.com/kermeow/static-opus v0.0.5

replace github.com/kermeow/rnnoise => github.com/kermeow/static-rnnoise v0.0.3

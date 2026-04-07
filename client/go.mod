module coop-voicechat

go 1.26.1

require (
	github.com/energye/systray v1.0.3
	github.com/gordonklaus/portaudio v0.0.0-20260203164431-765aa7dfa631
	github.com/hraban/opus v0.0.0-20251117090126-c76ea7e21bf3
	github.com/postfinance/single v0.0.2
	github.com/quartercastle/vector v0.2.0
)

require github.com/godbus/dbus/v5 v5.1.0 // indirect

replace github.com/gordonklaus/portaudio => github.com/KarpelesLab/static-portaudio v0.6.190600

replace github.com/hraban/opus => github.com/KarpelesLab/static-opus v0.9.152

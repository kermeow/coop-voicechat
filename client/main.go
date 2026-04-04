package main

import (
	"coop-voicechat/coop"
	"coop-voicechat/ui"
	"log"
	"os"

	"gioui.org/app"
	"github.com/gordonklaus/portaudio"
	// "gopkg.in/hraban/opus.v2"
)

func main() {
	err := portaudio.Initialize()
	if err != nil {
		panic(err)
	}
	defer portaudio.Terminate()

	coop.EnsureDirs()

	go func() {
		ui := ui.New()
		if err := ui.Run(); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	app.Main()
}

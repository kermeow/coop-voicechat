package main

import (
	"coop-voicechat/coop"
	"log"

	"github.com/energye/systray"
	"github.com/gordonklaus/portaudio"
	"github.com/postfinance/single"
)

var bridge *coop.Bridge

func main() {
	one, err := single.New("coop-voicechat")
	if err != nil {
		log.Fatalln(err)
	}
	if err := one.Lock(); err != nil {
		log.Fatalln(err)
	}
	defer one.Unlock()

	log.Println("Initialize PortAudio")
	err = portaudio.Initialize()
	if err != nil {
		panic(err)
	}
	defer portaudio.Terminate()

	log.Println("Checking sm64coopdx dirs")
	coop.EnsureDirs()

	bridge = coop.NewBridge()
	go bridge.Run()
	defer bridge.Stop()

	systray.Run(onReady, onExit)
	log.Println("Bye bye!")
}

func onReady() {
	systray.SetTitle("coop-voicechat")
	systray.SetTooltip("coop-voicechat client")

	systray.SetOnClick(func(menu systray.IMenu) {
		if menu != nil {
			menu.ShowMenu()
		}
	})

	mStatus := systray.AddMenuItem("Disconnected", "Current bridge status")
	mStatus.Disable()
	go func() {
		for bridge.Running {
			switch e := <-bridge.Event; e {
			case coop.BridgeConnect:
				mStatus.SetTitle("Connected")
			case coop.BridgeDisconnect:
				mStatus.SetTitle("Disconnected")
			}
		}
	}()

	systray.AddSeparator()

	mQuit := systray.AddMenuItem("Quit", "Quit the coop-voicechat client")
	mQuit.Click(systray.Quit)
}

func onExit() {}

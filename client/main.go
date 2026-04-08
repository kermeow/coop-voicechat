package main

import (
	"coop-voicechat/assets"
	"coop-voicechat/config"
	"coop-voicechat/coop"
	"log"

	"github.com/energye/systray"
	"github.com/gordonklaus/portaudio"
	"github.com/postfinance/single"
)

var GitDescribe string = "unknown"
var GitBranch string = "unknown"
var GitCommit string = "unknown"

var bridge *coop.Bridge
var options *config.Config

func main() {
	log.Printf("coop-voicechat %s (%s@%s)\n", GitDescribe, GitCommit, GitBranch)

	one, err := single.New("coop-voicechat")
	if err != nil {
		log.Fatalln(err)
	}
	if err := one.Lock(); err != nil {
		log.Fatalln(err)
	}
	defer one.Unlock()

	log.Println("Reading options")
	options, _ = config.Load(coop.VoiceOptions)
	defer options.Save(coop.VoiceOptions)

	log.Println("Initialize PortAudio")
	err = portaudio.Initialize()
	if err != nil {
		panic(err)
	}
	defer portaudio.Terminate()

	log.Println("Checking sm64coopdx dirs")
	coop.EnsureDirs()

	bridge = coop.NewBridge(options)
	go bridge.Run()
	defer bridge.Stop()

	systray.Run(onReady, onExit)
	log.Println("Bye bye!")
}

func onReady() {
	systray.SetIcon(assets.Microphone)
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

	mPanning := systray.AddMenuItemCheckbox("Stereo Panning", "Hear players to your left and right", options.StereoPanning)
	mPanning.Click(handleCheckbox(&options.StereoPanning, mPanning))

	systray.AddSeparator()

	mQuit := systray.AddMenuItem("Quit", "Quit the coop-voicechat client")
	mQuit.Click(systray.Quit)
}

func onExit() {}

func handleCheckbox(b *bool, m *systray.MenuItem) func() {
	return func() {
		*b = !*b
		if *b {
			m.Check()
		} else {
			m.Uncheck()
		}
	}
}

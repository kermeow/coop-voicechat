package main

import (
	"context"
	"coop-voicechat/assets"
	"coop-voicechat/audio"
	"coop-voicechat/bridge"
	"coop-voicechat/paths"
	"log"
	"time"

	"github.com/energye/systray"
	"github.com/gopxl/beep/v2"
	"github.com/gopxl/beep/v2/speaker"
	"github.com/gordonklaus/portaudio"
	"github.com/postfinance/single"
)

var GitDescribe string = "unknown"
var GitBranch string = "unknown"
var GitCommit string = "unknown"

var gVoiceBridge *bridge.Bridge

func main() {
	log.Printf("coop-voicechat %s (%s@%s)\n", GitDescribe, GitCommit, GitBranch)
	defer log.Println("Bye bye!")

	one, err := single.New("coop-voicechat")
	if err != nil {
		log.Fatalln(err)
	}
	if err := one.Lock(); err != nil {
		log.Fatalln(err)
	}
	defer one.Unlock()

	ctx, cancel := context.WithCancel(context.Background())

	log.Println("Checking sm64coopdx dirs")
	paths.EnsureDirs()

	log.Println("Initialize PortAudio")
	err = portaudio.Initialize()
	if err != nil {
		log.Fatalln(err)
		return
	}
	defer portaudio.Terminate()

	log.Println("Initialize speaker")
	sr := beep.SampleRate(audio.SAMPLE_RATE)
	speaker.Init(sr, sr.N(50*time.Millisecond))
	defer speaker.Clear()

	gVoiceBridge = bridge.NewBridge()

	go gVoiceBridge.Run(ctx)
	
	systray.Run(onReady, onExit)
	cancel()
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

	systray.AddSeparator()

	// TODO: allow input device changing

	// mInputDevice := systray.AddMenuItem("Input Device", "Change input device")
	// mInputDevices := make(map[string]systray.MenuItem)

	// mOutputDevice := systray.AddMenuItem("Output Device", "Change output device")
	// mOutputDevices := make(map[string]systray.MenuItem)

	// allDevices, _ := portaudio.Devices()
	// inputDefault, _ := portaudio.DefaultInputDevice()
	// outputDefault, _ := portaudio.DefaultOutputDevice()

	// for _, device := range allDevices {
	// 	if device.MaxInputChannels > 0 {
	// 		mDevice := mInputDevice.AddSubMenuItemCheckbox(device.Name, "", device == inputDefault)
	// 		mInputDevices[device.Name] = *mDevice
	// 	}
	// 	if device.MaxOutputChannels > 0 {
	// 		mDevice := mOutputDevice.AddSubMenuItemCheckbox(device.Name, "", device == outputDefault)
	// 		mOutputDevices[device.Name] = *mDevice
	// 	}
	// }

	systray.AddSeparator()

	mQuit := systray.AddMenuItem("Quit", "Quit the coop-voicechat client")
	mQuit.Click(systray.Quit)

	go func() {
		for e := range gVoiceBridge.Event {
			switch e {
			case bridge.BridgeConnect:
				mStatus.SetTitle("Connected")
			case bridge.BridgeDisconnect:
				mStatus.SetTitle("Disconnected")
			default:
			}
		}
	}()
}

func onExit() {
}

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

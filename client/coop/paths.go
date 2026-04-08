package coop

import (
	"log"
	"os"
	"path"
	"runtime"
)

func ensureDir(dir string) {
	if stat, err := os.Stat(dir); err != nil || !stat.IsDir() {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			panic(err)
		}
		log.Println("Made dir", dir)
		return
	}
	log.Println("Dir exists", dir)
}

func findAppData() string {
	appdata, err := os.UserConfigDir()
	if err != nil {
		panic(err)
	}

	if runtime.GOOS == "linux" {
		home, err := os.UserHomeDir()
		if err != nil {
			panic(err)
		}
		appdata = path.Join(home, ".local", "share")
	}

	if stat, err := os.Stat(path.Join(appdata, "sm64ex-coop")); err == nil && stat.IsDir() {
		return path.Join(appdata, "sm64ex-coop")
	}
	return path.Join(appdata, "sm64coopdx")
}

var (
	GameDir      = findAppData()
	SavDir       = path.Join(GameDir, "sav")
	
	VoiceOptions = path.Join(GameDir, "coop-voicechat.toml")
)

func EnsureDirs() {
	ensureDir(GameDir)
	ensureDir(SavDir)
}

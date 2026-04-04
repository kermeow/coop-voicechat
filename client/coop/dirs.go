package coop

import (
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
	}
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
	AppData = findAppData()
	Sav     = path.Join(AppData, "sav")
)

func EnsureDirs() {
	ensureDir(AppData)
	ensureDir(Sav)
}

//go:build linux

package paths

import (
	"log"
	"os"
	"path"
)

func findAppData() string {
	appdata, err := os.UserHomeDir()
	if err != nil {
		log.Fatalln(err)
	}
	appdata = path.Join(appdata, ".local", "share")

	if stat, err := os.Stat(path.Join(appdata, "sm64ex-coop")); err == nil && stat.IsDir() {
		return path.Join(appdata, "sm64ex-coop")
	}
	return path.Join(appdata, "sm64coopdx")
}

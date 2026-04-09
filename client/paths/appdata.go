//go:build !linux

package paths

import (
	"log"
	"os"
	"path"
)

func findAppData() string {
	appdata, err := os.UserConfigDir()
	if err != nil {
		log.Fatalln(err)
	}

	if stat, err := os.Stat(path.Join(appdata, "sm64ex-coop")); err == nil && stat.IsDir() {
		return path.Join(appdata, "sm64ex-coop")
	}
	return path.Join(appdata, "sm64coopdx")
}

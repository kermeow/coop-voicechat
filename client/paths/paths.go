package paths

import (
	"log"
	"os"
	"path"
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

var (
	GameDir = findAppData()
	SavDir  = path.Join(GameDir, "sav")

	// VoiceOptions = path.Join(GameDir, "coop-voicechat.toml")
)

func EnsureDirs() {
	ensureDir(GameDir)
	ensureDir(SavDir)
}

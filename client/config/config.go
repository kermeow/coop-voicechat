package config

import (
	"os"

	"github.com/BurntSushi/toml"
)

type Config struct {
	StereoPanning bool `toml:"stereo_panning"`
}

func Default() *Config {
	return &Config{
		StereoPanning: true,
	}
}

func Load(fp string) (*Config, error) {
	d := Default()

	_, err := os.Stat(fp)
	if err != nil {
		return d, err
	}

	_, err = toml.DecodeFile(fp, &d)
	return d, err
}

func (c *Config) Save(fp string) error {
	f, err := os.OpenFile(fp, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	defer f.Close()

	e := toml.NewEncoder(f)
	err = e.Encode(c)
	return err
}

package ymp3d

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Server ServerConfig
	Log    LogConfig
}

type ServerConfig struct {
	IP          string `toml:"IP"`
	Port        string `toml:"Port"`
	DownloadDir string `toml:"DownloadDir"`
}

type LogConfig struct {
	File  string `toml:"File"`
	Level string `toml:"Level"`
}

func newConfig() (c *Config) {
	const configPath string = "/etc/ymp3d.tml"
	c = new(Config)
	_, err := toml.DecodeFile(configPath, c)
	if err != nil {
		fmt.Printf("No exist %s or Fail to Decode %s", configPath, configPath)
		panic(err)
	}
	return c
}

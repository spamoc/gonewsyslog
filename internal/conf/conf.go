package conf

import "github.com/BurntSushi/toml"

func New(path string) (Config, error) {
	var c Config
	_, err := toml.DecodeFile(path, &c)
	if err != nil {
		return Config{}, err
	}
	return c, nil
}

type Config struct {
	Common   CommonConfig
	Projects []ProjectConfig `toml:"project"`
}

type CommonConfig struct {
	Include string `toml:"include"`
}

type archiveType string

// if you change it, should change syslog/initializer const, too.
const (
	GZIP archiveType = "gzip"
	BZ2              = "bz2"
	XZ               = "xz"
	NONE             = "none"
)

type ProjectConfig struct {
	Name     string         `toml:"name"`
	From     string         `toml:"from"`
	To       string         `toml:"to"`
	Rotate   RotateConfig   `toml:"rotate"`
	Compress CompressConfig `toml:"compress"`
	Success  string         `toml:"successCommand"`
	Failed   string         `toml:"failedCommand"`
	Pid      string         `toml:"pid"`
}

type RotateConfig struct {
	Term  string `toml:"term"`
	Size  string `toml:"size"`
	Count int    `toml:"count"`
}

type CompressConfig struct {
	Type archiveType `toml:"type"`
	Ext  string      `toml:"ext"`
}

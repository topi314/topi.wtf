package topi

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

func LoadConfig(path string) (Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return Config{}, err
	}

	var cfg Config
	if err = yaml.NewDecoder(file).Decode(&cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

type Config struct {
	Log        LogConfig    `yaml:"log"`
	Debug      bool         `yaml:"debug"`
	DevMode    bool         `yaml:"dev_mode"`
	ListenAddr string       `yaml:"listen_addr"`
	GitHub     GitHubConfig `yaml:"github"`
	Cache      *CacheConfig `yaml:"cache"`
	LastFM     LastFMConfig `yaml:"lastfm"`
}

func (c Config) String() string {
	return fmt.Sprintf("\n Log: %s\n DevMode: %t\n Debug: %t\n ListenAddr: %s\n GitHub: %s\n Cache: %s\n LastFM: %s\n",
		c.Log,
		c.DevMode,
		c.Debug,
		c.ListenAddr,
		c.GitHub,
		c.Cache,
		c.LastFM,
	)
}

type LogConfig struct {
	Level     slog.Level `cfg:"level"`
	Format    string     `cfg:"format"`
	AddSource bool       `cfg:"add_source"`
	NoColor   bool       `cfg:"no_color"`
}

func (c LogConfig) String() string {
	return fmt.Sprintf("\n  Level: %s\n  Format: %s\n  AddSource: %t\n  NoColor: %t\n",
		c.Level,
		c.Format,
		c.AddSource,
		c.NoColor,
	)
}

type GitHubConfig struct {
	AccessToken string `yaml:"access_token"`
	User        string `yaml:"user"`
}

func (c GitHubConfig) String() string {
	return fmt.Sprintf("\n  AccessToken: %s\n  User: %s",
		strings.Repeat("*", len(c.AccessToken)),
		c.User,
	)
}

type CacheConfig struct {
	Size int           `yaml:"size"`
	TTL  time.Duration `yaml:"ttl"`
}

func (c CacheConfig) String() string {
	return fmt.Sprintf("\n  Size: %d\n  TTL: %s",
		c.Size,
		c.TTL,
	)
}

type LastFMConfig struct {
	Username string        `yaml:"username"`
	APIKey   string        `yaml:"api_key"`
	Size     int           `yaml:"size"`
	TTL      time.Duration `yaml:"ttl"`
}

func (c LastFMConfig) String() string {
	return fmt.Sprintf("\n  Username: %s\n  APIKey: %s\n  Size: %d\n  TTL: %s",
		c.Username,
		strings.Repeat("*", len(c.APIKey)),
		c.Size,
		c.TTL,
	)
}

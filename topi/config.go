package topi

import (
	"fmt"
	"strings"
	"time"
)

type Config struct {
	Debug      bool         `cfg:"debug"`
	DevMode    bool         `cfg:"dev_mode"`
	ListenAddr string       `cfg:"listen_addr"`
	Blog       BlogConfig   `cfg:"blog"`
	GitHub     GitHubConfig `cfg:"github"`
	Cache      *CacheConfig `cfg:"cache"`
}

func (c Config) String() string {
	return fmt.Sprintf("\n DevMode: %t\n Debug: %t\n ListenAddr: %s\n GitHub: %s\n Cache: %s\n", c.DevMode, c.Debug, c.ListenAddr, c.GitHub, c.Cache)
}

type BlogConfig struct {
	Repository string `cfg:"repository"`
	User       string `cfg:"user"`
	Category   string `cfg:"category"`
}

func (c BlogConfig) String() string {
	return fmt.Sprintf("\n   Repository: %s\n   User: %s\n   Category: %s", c.Repository, c.User, c.Category)
}

type GitHubConfig struct {
	AccessToken  string `cfg:"access_token"`
	ClientID     string `cfg:"client_id"`
	ClientSecret string `cfg:"client_secret"`
	RedirectURL  string `cfg:"redirect_url"`
}

func (c GitHubConfig) String() string {
	return fmt.Sprintf("\n  AccessToken: %s\n  ClientID: %s\n  ClientSecret: %s\n  RedirectURL: %s", strings.Repeat("*", len(c.AccessToken)), c.ClientID, strings.Repeat("*", len(c.ClientSecret)), c.RedirectURL)
}

type CacheConfig struct {
	Size int           `cfg:"size"`
	TTL  time.Duration `cfg:"ttl"`
}

func (c CacheConfig) String() string {
	return fmt.Sprintf("\n  Size: %d\n  TTL: %s\n", c.Size, c.TTL)
}

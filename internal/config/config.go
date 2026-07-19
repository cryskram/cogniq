package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Core    CoreConfig    `mapstructure:"core"`
	Daemon  DaemonConfig  `mapstructure:"daemon"`
	MCP     MCPConfig     `mapstructure:"mcp"`
	Indexer IndexerConfig `mapstructure:"indexer"`
	Watcher WatcherConfig `mapstructure:"watcher"`
	Search  SearchConfig  `mapstructure:"search"`
	Log     LogConfig     `mapstructure:"log"`
}

type CoreConfig struct {
	DataDir string `mapstructure:"data_dir"`
}

type DaemonConfig struct {
	Socket  string `mapstructure:"socket"`
	TCPHost string `mapstructure:"tcp_host"`
	TCPPort int    `mapstructure:"tcp_port"`
}

type MCPConfig struct {
	Enabled   bool   `mapstructure:"enabled"`
	Transport string `mapstructure:"transport"`
	TCPPort   int    `mapstructure:"tcp_port"`
}

type IndexerConfig struct {
	Concurrency int   `mapstructure:"concurrency"`
	MaxFileSize int64 `mapstructure:"max_file_size"`
	MaxCommits  int   `mapstructure:"max_commits"`
}

type WatcherConfig struct {
	Enabled  bool          `mapstructure:"enabled"`
	Debounce time.Duration `mapstructure:"debounce"`
}

type SearchConfig struct {
	MaxResults   int  `mapstructure:"max_results"`
	PathBoosting bool `mapstructure:"path_boosting"`
}

type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
	Output string `mapstructure:"output"`
}

func Load() (*Config, error) {
	v := viper.New()

	dataDir, err := DefaultDataDir()
	if err != nil {
		return nil, fmt.Errorf("default data dir: %w", err)
	}
	v.SetDefault("core.data_dir", dataDir)

	socketPath, err := DefaultSocketPath()
	if err != nil {
		return nil, fmt.Errorf("default socket path: %w", err)
	}
	v.SetDefault("daemon.socket", socketPath)
	v.SetDefault("daemon.tcp_host", "127.0.0.1")
	v.SetDefault("daemon.tcp_port", 9876)

	v.SetDefault("mcp.enabled", true)
	v.SetDefault("mcp.transport", "stdio")
	v.SetDefault("mcp.tcp_port", 9877)

	v.SetDefault("indexer.concurrency", 4)
	v.SetDefault("indexer.max_file_size", 10*1024*1024)
	v.SetDefault("indexer.max_commits", 10000)

	v.SetDefault("watcher.enabled", true)
	v.SetDefault("watcher.debounce", time.Second)

	v.SetDefault("search.max_results", 100)
	v.SetDefault("search.path_boosting", true)

	v.SetDefault("log.level", "info")
	v.SetDefault("log.format", "console")
	v.SetDefault("log.output", "stderr")

	v.SetEnvPrefix("COGNIQ")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	var cfg Config

	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
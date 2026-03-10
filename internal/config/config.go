package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	General   GeneralConfig   `yaml:"general"`
	Contexts  []Context       `yaml:"contexts"`
	Dashboard DashboardConfig `yaml:"dashboard"`
	LogQuery  LogQueryConfig  `yaml:"log_query"`
}

type GeneralConfig struct {
	TimeZone string `yaml:"time_zone"`
	Theme    string `yaml:"theme"`
	Region   string `yaml:"region"`
}

type DashboardConfig struct {
	AutoPoll            bool   `yaml:"auto_poll"`
	PollIntervalSeconds int    `yaml:"poll_interval_seconds"`
	AggregationInterval string `yaml:"aggregation_interval"`
	BufferSize          int    `yaml:"buffer_size"`
}

type LogQueryConfig struct {
	SavedQueries []SavedQuery `yaml:"saved_queries"`
}

type SavedQuery struct {
	Name  string `yaml:"name"`
	Query string `yaml:"query"`
}

func DefaultConfig() *Config {
	return &Config{
		General: GeneralConfig{
			TimeZone: "UTC",
			Theme:    "dark",
		},
		Dashboard: DashboardConfig{
			PollIntervalSeconds: 5,
			AggregationInterval: "1m",
			BufferSize:          10000,
		},
	}
}

func configDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "xatu")
}

func configPath() string {
	return filepath.Join(configDir(), "config.yaml")
}

func Exists() bool {
	_, err := os.Stat(configPath())
	return err == nil
}

func Load() (*Config, error) {
	cfg := DefaultConfig()

	data, err := os.ReadFile(configPath())
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, err
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	// Environment overrides
	if region := os.Getenv("AWS_REGION"); region != "" && cfg.General.Region == "" {
		cfg.General.Region = region
	}
	if region := os.Getenv("AWS_DEFAULT_REGION"); region != "" && cfg.General.Region == "" {
		cfg.General.Region = region
	}

	return cfg, nil
}

func Save(cfg *Config) error {
	dir := configDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(configPath(), data, 0644)
}

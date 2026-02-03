// Package config handles loading and validating server configuration from YAML files.
package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config holds the complete server configuration.
type Config struct {
	Server  ServerConfig  `yaml:"server"`
	User    UserConfig    `yaml:"user"`
	Storage StorageConfig `yaml:"storage"`
	Logging LoggingConfig `yaml:"logging"`
}

// ServerConfig holds HTTP server settings.
type ServerConfig struct {
	Host        string `yaml:"host"`
	Port        int    `yaml:"port"`
	ExternalURL string `yaml:"external_url"`
}

// UserConfig holds the preconfigured user credentials.
type UserConfig struct {
	Email    string `yaml:"email"`
	Password string `yaml:"password"`
}

// StorageConfig holds database and content storage settings.
type StorageConfig struct {
	DBPath     string `yaml:"db_path"`
	ContentDir string `yaml:"content_dir"`
	QuotaBytes int64  `yaml:"quota_bytes"`
}

// LoggingConfig holds logging settings.
type LoggingConfig struct {
	Level string `yaml:"level"`
}

// Load reads and parses a YAML configuration file from the given path.
// Returns an error if the file cannot be read or parsed.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parsing config file: %w", err)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return cfg, nil
}

// Addr returns the listen address in "host:port" format.
func (c *Config) Addr() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}

// validate checks that required configuration fields are present and valid.
func (c *Config) validate() error {
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("server.port must be between 1 and 65535")
	}
	if c.Server.ExternalURL == "" {
		return fmt.Errorf("server.external_url is required")
	}
	if c.User.Email == "" {
		return fmt.Errorf("user.email is required")
	}
	if c.User.Password == "" {
		return fmt.Errorf("user.password is required")
	}
	if c.Storage.DBPath == "" {
		return fmt.Errorf("storage.db_path is required")
	}
	if c.Storage.ContentDir == "" {
		return fmt.Errorf("storage.content_dir is required")
	}
	if c.Storage.QuotaBytes <= 0 {
		return fmt.Errorf("storage.quota_bytes must be positive")
	}
	return nil
}

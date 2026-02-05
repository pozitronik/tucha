// Package config handles loading and validating server configuration from YAML files.
package config

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config holds the complete server configuration.
type Config struct {
	Server    ServerConfig    `yaml:"server"`
	Admin     AdminConfig     `yaml:"admin"`
	Storage   StorageConfig   `yaml:"storage"`
	Auth      AuthConfig      `yaml:"auth"`
	Logging   LoggingConfig   `yaml:"logging"`
	Endpoints EndpointsConfig `yaml:"endpoints"`
}

// ServerConfig holds HTTP server settings.
type ServerConfig struct {
	Host        string `yaml:"host"`
	Port        int    `yaml:"port"`
	ExternalURL string `yaml:"external_url"`
}

// AdminConfig holds the admin panel credentials (not stored in DB).
type AdminConfig struct {
	Login    string `yaml:"login"`
	Password string `yaml:"password"`
}

// StorageConfig holds database and content storage settings.
type StorageConfig struct {
	DBPath     string `yaml:"db_path"`
	ContentDir string `yaml:"content_dir"`
	QuotaBytes int64  `yaml:"quota_bytes"`
}

// AuthConfig holds authentication settings.
type AuthConfig struct {
	TokenTTLSeconds int `yaml:"token_ttl_seconds"`
}

// LoggingConfig holds logging settings.
type LoggingConfig struct {
	Level  string `yaml:"level"`  // DEBUG, INFO, WARN, ERROR (default: INFO)
	Output string `yaml:"output"` // stdout, file, both (default: stdout)
	File   string `yaml:"file"`   // Path when output is "file" or "both"
}

// EndpointsConfig holds per-endpoint URLs for service discovery.
// Each field is optional; unset values are derived from server.external_url.
type EndpointsConfig struct {
	API        string `yaml:"api"`
	OAuth      string `yaml:"oauth"`
	Dispatcher string `yaml:"dispatcher"`
	Upload     string `yaml:"upload"`
	Download   string `yaml:"download"`
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

	cfg.applyDefaults()

	return cfg, nil
}

// Addr returns the listen address in "host:port" format.
func (c *Config) Addr() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}

// applyDefaults fills in unset configuration values with sensible defaults.
func (c *Config) applyDefaults() {
	// Auth defaults
	if c.Auth.TokenTTLSeconds <= 0 {
		c.Auth.TokenTTLSeconds = 86400 // 24 hours
	}

	// Logging defaults
	if c.Logging.Level == "" {
		c.Logging.Level = "info"
	}
	if c.Logging.Output == "" {
		c.Logging.Output = "stdout"
	}

	// Endpoint defaults
	base := strings.TrimRight(c.Server.ExternalURL, "/")
	if c.Endpoints.API == "" {
		c.Endpoints.API = base + "/api/v2"
	}
	if c.Endpoints.OAuth == "" {
		c.Endpoints.OAuth = base + "/token"
	}
	if c.Endpoints.Dispatcher == "" {
		c.Endpoints.Dispatcher = base + "/api/v2/dispatcher"
	}
	if c.Endpoints.Upload == "" {
		c.Endpoints.Upload = base + "/upload"
	}
	if c.Endpoints.Download == "" {
		c.Endpoints.Download = base + "/get"
	}
}

// validate checks that required configuration fields are present and valid.
func (c *Config) validate() error {
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("server.port must be between 1 and 65535")
	}
	if c.Server.ExternalURL == "" {
		return fmt.Errorf("server.external_url is required")
	}
	if c.Admin.Login == "" {
		return fmt.Errorf("admin.login is required")
	}
	if c.Admin.Password == "" {
		return fmt.Errorf("admin.password is required")
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

	// Logging validation: file path required for file/both output modes
	output := strings.ToLower(c.Logging.Output)
	if (output == "file" || output == "both") && c.Logging.File == "" {
		return fmt.Errorf("logging.file is required when logging.output is %q", c.Logging.Output)
	}

	return nil
}

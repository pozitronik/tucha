package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeConfig(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

const validYAML = `
server:
  host: "0.0.0.0"
  port: 8080
  external_url: "http://localhost:8080"
admin:
  login: "admin"
  password: "secret"
storage:
  db_path: "/tmp/test.db"
  content_dir: "/tmp/content"
  quota_bytes: 1073741824
logging:
  level: "info"
`

func TestLoad_valid(t *testing.T) {
	p := writeConfig(t, validYAML)
	cfg, err := Load(p)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Server.Port != 8080 {
		t.Errorf("port = %d, want 8080", cfg.Server.Port)
	}
	if cfg.Admin.Login != "admin" {
		t.Errorf("login = %q, want %q", cfg.Admin.Login, "admin")
	}
	if cfg.Storage.QuotaBytes != 1073741824 {
		t.Errorf("quota = %d, want 1073741824", cfg.Storage.QuotaBytes)
	}
}

func TestLoad_missingFile(t *testing.T) {
	_, err := Load("/nonexistent/path/config.yaml")
	if err == nil {
		t.Error("Load with missing file should return error")
	}
}

func TestLoad_invalidYAML(t *testing.T) {
	p := writeConfig(t, "{{invalid yaml")
	_, err := Load(p)
	if err == nil {
		t.Error("Load with invalid YAML should return error")
	}
}

func TestLoad_validation(t *testing.T) {
	tests := []struct {
		name      string
		override  string
		errSubstr string
	}{
		{
			"zero port",
			`server: { host: "", port: 0, external_url: "http://x" }
admin: { login: "a", password: "b" }
storage: { db_path: "x", content_dir: "y", quota_bytes: 1 }`,
			"port",
		},
		{
			"port too high",
			`server: { host: "", port: 70000, external_url: "http://x" }
admin: { login: "a", password: "b" }
storage: { db_path: "x", content_dir: "y", quota_bytes: 1 }`,
			"port",
		},
		{
			"missing external_url",
			`server: { host: "", port: 8080, external_url: "" }
admin: { login: "a", password: "b" }
storage: { db_path: "x", content_dir: "y", quota_bytes: 1 }`,
			"external_url",
		},
		{
			"missing admin login",
			`server: { host: "", port: 8080, external_url: "http://x" }
admin: { login: "", password: "b" }
storage: { db_path: "x", content_dir: "y", quota_bytes: 1 }`,
			"admin.login",
		},
		{
			"missing admin password",
			`server: { host: "", port: 8080, external_url: "http://x" }
admin: { login: "a", password: "" }
storage: { db_path: "x", content_dir: "y", quota_bytes: 1 }`,
			"admin.password",
		},
		{
			"missing db_path",
			`server: { host: "", port: 8080, external_url: "http://x" }
admin: { login: "a", password: "b" }
storage: { db_path: "", content_dir: "y", quota_bytes: 1 }`,
			"db_path",
		},
		{
			"missing content_dir",
			`server: { host: "", port: 8080, external_url: "http://x" }
admin: { login: "a", password: "b" }
storage: { db_path: "x", content_dir: "", quota_bytes: 1 }`,
			"content_dir",
		},
		{
			"zero quota",
			`server: { host: "", port: 8080, external_url: "http://x" }
admin: { login: "a", password: "b" }
storage: { db_path: "x", content_dir: "y", quota_bytes: 0 }`,
			"quota_bytes",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := writeConfig(t, tt.override)
			_, err := Load(p)
			if err == nil {
				t.Fatal("expected validation error")
			}
			if !strings.Contains(err.Error(), tt.errSubstr) {
				t.Errorf("error %q does not contain %q", err.Error(), tt.errSubstr)
			}
		})
	}
}

func TestConfig_Addr(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{Host: "127.0.0.1", Port: 9090},
	}
	if got := cfg.Addr(); got != "127.0.0.1:9090" {
		t.Errorf("Addr() = %q, want %q", got, "127.0.0.1:9090")
	}
}

func TestLoad_endpointDefaults(t *testing.T) {
	p := writeConfig(t, validYAML)
	cfg, err := Load(p)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	base := "http://localhost:8080"
	if cfg.Endpoints.API != base+"/api/v2" {
		t.Errorf("API = %q, want %q", cfg.Endpoints.API, base+"/api/v2")
	}
	if cfg.Endpoints.OAuth != base+"/token" {
		t.Errorf("OAuth = %q, want %q", cfg.Endpoints.OAuth, base+"/token")
	}
	if cfg.Endpoints.Dispatcher != base+"/api/v2/dispatcher" {
		t.Errorf("Dispatcher = %q, want %q", cfg.Endpoints.Dispatcher, base+"/api/v2/dispatcher")
	}
	if cfg.Endpoints.Upload != base+"/upload" {
		t.Errorf("Upload = %q, want %q", cfg.Endpoints.Upload, base+"/upload")
	}
	if cfg.Endpoints.Download != base+"/get" {
		t.Errorf("Download = %q, want %q", cfg.Endpoints.Download, base+"/get")
	}
}

func TestLoad_trailingSlashTrimmed(t *testing.T) {
	yaml := `
server:
  host: "0.0.0.0"
  port: 8080
  external_url: "http://localhost:8080/"
admin:
  login: "admin"
  password: "secret"
storage:
  db_path: "/tmp/test.db"
  content_dir: "/tmp/content"
  quota_bytes: 1073741824
`
	p := writeConfig(t, yaml)
	cfg, err := Load(p)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	// Endpoints should not have double slashes from trailing slash in external_url.
	if strings.Contains(cfg.Endpoints.API, "//api") {
		t.Errorf("API endpoint has double slash: %q", cfg.Endpoints.API)
	}
}

func TestLoad_partialEndpointOverride(t *testing.T) {
	yaml := `
server:
  host: "0.0.0.0"
  port: 8080
  external_url: "http://localhost:8080"
admin:
  login: "admin"
  password: "secret"
storage:
  db_path: "/tmp/test.db"
  content_dir: "/tmp/content"
  quota_bytes: 1073741824
endpoints:
  api: "http://custom:9090/api/v2"
`
	p := writeConfig(t, yaml)
	cfg, err := Load(p)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if cfg.Endpoints.API != "http://custom:9090/api/v2" {
		t.Errorf("overridden API = %q", cfg.Endpoints.API)
	}
	// Non-overridden endpoints should use default base.
	if cfg.Endpoints.OAuth != "http://localhost:8080/token" {
		t.Errorf("default OAuth = %q", cfg.Endpoints.OAuth)
	}
}

func TestLoad_loggingDefaults(t *testing.T) {
	p := writeConfig(t, validYAML)
	cfg, err := Load(p)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if cfg.Logging.Level != "info" {
		t.Errorf("Logging.Level = %q, want %q", cfg.Logging.Level, "info")
	}
	if cfg.Logging.Output != "stdout" {
		t.Errorf("Logging.Output = %q, want %q", cfg.Logging.Output, "stdout")
	}
}

func TestLoad_loggingFileRequired(t *testing.T) {
	tests := []struct {
		name   string
		output string
	}{
		{"file output", "file"},
		{"both output", "both"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			yaml := `
server:
  host: "0.0.0.0"
  port: 8080
  external_url: "http://localhost:8080"
admin:
  login: "admin"
  password: "secret"
storage:
  db_path: "/tmp/test.db"
  content_dir: "/tmp/content"
  quota_bytes: 1073741824
logging:
  level: "info"
  output: "` + tt.output + `"
`
			p := writeConfig(t, yaml)
			_, err := Load(p)
			if err == nil {
				t.Error("expected validation error for missing logging.file")
			}
			if !strings.Contains(err.Error(), "logging.file") {
				t.Errorf("error %q does not contain 'logging.file'", err.Error())
			}
		})
	}
}

func TestLoad_loggingWithFile(t *testing.T) {
	yaml := `
server:
  host: "0.0.0.0"
  port: 8080
  external_url: "http://localhost:8080"
admin:
  login: "admin"
  password: "secret"
storage:
  db_path: "/tmp/test.db"
  content_dir: "/tmp/content"
  quota_bytes: 1073741824
logging:
  level: "debug"
  output: "both"
  file: "/tmp/tucha.log"
`
	p := writeConfig(t, yaml)
	cfg, err := Load(p)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if cfg.Logging.Level != "debug" {
		t.Errorf("Logging.Level = %q, want %q", cfg.Logging.Level, "debug")
	}
	if cfg.Logging.Output != "both" {
		t.Errorf("Logging.Output = %q, want %q", cfg.Logging.Output, "both")
	}
	if cfg.Logging.File != "/tmp/tucha.log" {
		t.Errorf("Logging.File = %q, want %q", cfg.Logging.File, "/tmp/tucha.log")
	}
}

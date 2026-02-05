# Tucha

[![CI](https://github.com/pozitronik/tucha/actions/workflows/ci.yml/badge.svg)](https://github.com/pozitronik/tucha/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/pozitronik/tucha/branch/master/graph/badge.svg)](https://codecov.io/gh/pozitronik/tucha)
[![Go Report Card](https://goreportcard.com/badge/github.com/pozitronik/tucha)](https://goreportcard.com/report/github.com/pozitronik/tucha)
[![License: GPL v3](https://img.shields.io/badge/License-GPLv3-blue.svg)](https://www.gnu.org/licenses/gpl-3.0)

[Русский](README.ru.md)

Open-source server implementing a cloud storage API v2 protocol, compatible with the cloud.mail.ru web API.

[API Specification](API_SPEC.md)

## Quick Start

### Prerequisites

- Go 1.24+

### Build and Run

```bash
go build -o tucha ./cmd/tucha
./tucha -config config.yaml
```

The server creates the database and storage directory automatically on first run.

### Command-Line Interface

```
Usage: tucha [options] [command]

Options:
  -config <path>     Path to configuration file (default: config.yaml)

Commands:
  --help, --?        Show help message
  --version          Show version and exit
  --background       Run server in background (daemon mode)
  --status           Show if server is running
  --stop             Stop background server
  --config-check     Validate configuration file

User Management:
  --user list [mask]               List users (optional email filter)
  --user add <email> <pwd> [quota] Add user (quota: "16GB", "512MB")
  --user remove <email>            Remove user
  --user pwd <email> <pwd>         Set password
  --user quota <email> <quota>     Set quota
  --user info <email>              Show user details
```

**Examples:**

```bash
tucha                              # Start in foreground
tucha --background                 # Start in background (daemon mode)
tucha --status                     # Check if server is running
tucha --stop                       # Stop background server
tucha --user add user@x.com pass 8GB  # Add user with 8GB quota
tucha --user list *@example.com    # List users matching pattern
tucha --user quota user@x.com 16GB # Update user quota
```

## Configuration

All settings are in `config.yaml`. There are no environment variable overrides.

```yaml
server:
  host: "0.0.0.0"                        # Bind address
  port: 8081                              # Listen port
  external_url: "http://localhost:8081"   # Public URL announced to clients
  # pid_file: "./tucha.pid"              # Optional: PID file path for daemon mode

admin:
  login: "admin"                          # Admin panel login
  password: "admin"                       # Admin panel password

storage:
  db_path: "./data/tucha.db"             # SQLite database file path
  content_dir: "./data/storage"          # Content-addressable file storage directory
  # thumbnail_dir: "./data/storage/thumbs" # Optional: thumbnail cache (default: content_dir/thumbs)
  quota_bytes: 17179869184               # Default user quota in bytes (16 GiB)

logging:
  level: "info"                          # Log level: debug, info, warn, error
  output: "stdout"                       # Output: stdout, file, both
  # file: "./logs/tucha.log"             # Required if output is "file" or "both"

# Optional: override individual endpoint URLs (derived from external_url by default)
# endpoints:
#   api: "http://localhost:8081/api/v2"
#   oauth: "http://localhost:8081"
#   dispatcher: "http://localhost:8081/api/v2/dispatcher"
#   upload: "http://localhost:8081/upload"
#   download: "http://localhost:8081/get"
```

### Configuration Notes

- **`admin.login` / `admin.password`** -- admin panel credentials. These are separate from user accounts and are used only for the web-based admin interface.
- **`storage.quota_bytes`** -- default quota assigned to newly created users when no explicit quota is provided. Changing this value affects only future users.
- **`storage.thumbnail_dir`** -- optional. Directory for caching image thumbnails. Defaults to `<content_dir>/thumbs`.
- **`server.pid_file`** -- optional. Path to the PID file for daemon mode. Defaults to `tucha.pid` in the same directory as the config file.
- **`logging.output`** -- where to send log output: `stdout` (default), `file`, or `both`. When using `file` or `both`, `logging.file` must be specified.
- **`endpoints.*`** -- optional. If omitted, derived from `external_url`. Set them explicitly when the server is behind a reverse proxy with different internal/external URLs.
- All paths (`db_path`, `content_dir`) are relative to the working directory unless absolute.
- Validated at startup: `port` must be 1--65535, `quota_bytes` must be positive, all required fields must be non-empty.

### First Access

1. Open the admin panel at `http://localhost:8081/admin`
2. Log in with the admin credentials from `config.yaml` (`admin.login` / `admin.password`)
3. Create user accounts through the admin panel
4. Connect the desktop client to `http://localhost:8081`

## Architecture

Clean architecture with DDD principles. The composition root is `cmd/tucha/main.go`.

```
cmd/tucha/                          Entry point, dependency wiring
internal/
  cli/                              Command-line interface parser and user commands
  config/                           YAML configuration loading and validation
  domain/
    entity/                         Core entities: User, Node, Token, Content, Share, TrashItem
    repository/                     Repository interfaces (ports)
    vo/                             Value objects: CloudPath, ContentHash, NodeType, AccessLevel, etc.
  application/
    port/                           Outbound port interfaces (ContentStorage, Hasher, Logger)
    service/                        Application services (use case orchestration)
  infrastructure/
    sqlite/                         SQLite repository implementations
    contentstore/                   Disk-based content-addressable storage
    hasher/                         mrCloud hash algorithm implementation
    logger/                         Leveled logging implementation
    thumbnail/                      Image thumbnail generator
  transport/
    httpapi/                        HTTP handlers, DTOs, routing, admin panel
```

### Layer Dependencies

```
transport -> application/service -> domain/entity + domain/repository
                                 -> application/port
infrastructure -> domain/repository (implements interfaces)
              -> application/port  (implements interfaces)
cmd/tucha -> all (composition root only)
```

## Authentication

The server has two independent authentication systems.

### User Authentication (OAuth2)

OAuth2 password grant flow for desktop client access. The server acts as both authorization server and resource server.

1. Client sends `POST /token` with form data: `client_id=cloud-win`, `grant_type=password`, `username=<email>`, `password=<password>`
2. Server returns `access_token`, `refresh_token`, `expires_in` (86400 seconds = 24 hours)
3. All API calls include `?access_token=<token>` as a query parameter
4. Tokens are 64-character random hex strings generated via `crypto/rand`
5. Expired tokens are rejected with status 403

### Admin Authentication

Config-based login/password with in-memory bearer tokens. Admin endpoints at `/admin/*` use this system. Admin credentials are set in `config.yaml` and are not stored in the database.

## Server Management

### Daemon Mode

Run the server in background with `--background`. The server:
- Writes a PID file (default: `tucha.pid` next to config file)
- Detaches from terminal
- Logs to file only (if configured)

Control commands:
- `tucha --status` -- check if server is running
- `tucha --stop` -- stop the background server

### Graceful Shutdown

The server handles SIGTERM and SIGINT signals gracefully:
- Stops accepting new connections
- Waits up to 30 seconds for in-flight requests to complete
- Closes database connections properly
- Removes PID file on exit

## Storage

### Content-Addressable File Storage

Files are stored by hash in a two-level sharded directory structure:

```
<content_dir>/<first2>/<chars3-4>/<full_hash_40chars>
```

Example: hash `C172C6E2FF47284FF33F348FEA7EECE532F6C051` is stored at:

```
./data/storage/C1/72/C172C6E2FF47284FF33F348FEA7EECE532F6C051
```

Identical file contents are stored once (deduplication via reference counting in the `contents` table).

### Hash Algorithm (mrCloud)

Two modes depending on file size:

- **Small files (< 21 bytes):** Content is zero-padded to 20 bytes and hex-encoded directly (no cryptographic hash).
- **Large files (>= 21 bytes):** `SHA1("mrCloud" + content + decimal_size_string)`, result uppercase hex.

### Database Schema (SQLite)

Six tables:

| Table      | Purpose                                                                                                           |
|------------|-------------------------------------------------------------------------------------------------------------------|
| `users`    | User accounts: id, email, password, is_admin, quota_bytes, created                                                |
| `nodes`    | Virtual filesystem: id, user_id, parent_id, name, home (full path), node_type, size, hash, mtime, rev, grev, tree |
| `contents` | Content registry: hash, size, ref_count, created                                                                  |
| `tokens`   | Auth tokens: id, user_id, access_token, refresh_token, csrf_token, expires_at                                     |
| `trash`    | Trashbin: id, user_id, original path, node type, hash, size, deletion metadata                                    |
| `shares`   | Folder sharing: id, owner, path, invitee email, access level, invite token, mount info                            |

Schema is created automatically. Migrations run at startup if needed.

## Quota Management

- Each user has a `quota_bytes` limit
- Usage is calculated as the sum of all file node sizes for the user
- Uploads are blocked when `used + file_size > quota`
- Default quota for new users comes from `storage.quota_bytes` in config
- Quota can be set per-user via the admin API (`quota_bytes` parameter on add/edit)
- Current usage is visible in the admin panel and via `GET /api/v2/user/space`

## Testing

```bash
go test ./internal/... -v -count=1
```

Race detector for concurrent tests:

```bash
go test ./internal/application/service/ -run TestAdminAuth_concurrent -race
```

## Dependencies

| Package              | Purpose                                 |
|----------------------|-----------------------------------------|
| `gopkg.in/yaml.v3`   | YAML configuration parsing              |
| `modernc.org/sqlite` | Pure-Go SQLite driver (no CGO required) |
| Standard library     | Everything else                         |

# License
[LICENSE: GNU GPL v3.0](LICENSE)

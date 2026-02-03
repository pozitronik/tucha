# Tucha

Go server implementing a cloud storage API v2 protocol, compatible with the cloud-win desktop client.

## Quick Start

### Prerequisites

- Go 1.24+

### Build and Run

```bash
go build -o tucha ./cmd/tucha
./tucha -config config.yaml
```

The server creates the database and storage directory automatically on first run.

### Default Access

After startup, log in with the credentials from `config.yaml` (`user.email` / `user.password`). The seed user is always an administrator.

Admin panel: `http://localhost:8080/admin`

## Configuration

All settings are in `config.yaml`. There are no environment variable overrides.

```yaml
server:
  host: "0.0.0.0"                        # Bind address
  port: 8080                              # Listen port
  external_url: "http://localhost:8080"   # Public URL announced to clients

user:
  email: "user@tucha.local"              # Seed admin email
  password: "apppassword"                # Seed admin password

storage:
  db_path: "./data/tucha.db"             # SQLite database file path
  content_dir: "./data/storage"          # Content-addressable file storage directory
  quota_bytes: 17179869184               # Default user quota in bytes (16 GiB)

logging:
  level: "info"                          # Log level

# Optional: override individual endpoint URLs (derived from external_url by default)
endpoints:
  api: ""
  oauth: ""
  dispatcher: ""
  upload: ""
  download: ""
```

### Configuration Notes

- **`storage.quota_bytes`** is the default quota assigned to newly created users when no explicit quota is provided. Changing this value affects only future users, not existing ones.
- **`user.email` / `user.password`** define the seed admin account. On every startup, this account is upserted (created or updated). Changing these values in the config and restarting will update the seed account credentials.
- **`endpoints.*`** are optional. If omitted, they are derived from `external_url`. Set them explicitly when the server is behind a reverse proxy with different internal/external URLs.
- All paths (`db_path`, `content_dir`) are relative to the working directory unless absolute.
- Validated at startup: `port` must be 1--65535, `quota_bytes` must be positive, all required fields must be non-empty.

## Architecture

Clean architecture with DDD principles. The composition root is `cmd/tucha/main.go`.

```
cmd/tucha/                          Entry point, dependency wiring
internal/
  config/                           YAML configuration loading and validation
  domain/
    entity/                         Core entities: User, Node, Token, Content
    repository/                     Repository interfaces (ports)
    vo/                             Value objects: CloudPath, ContentHash, NodeType, ConflictMode
  application/
    port/                           Outbound port interfaces (ContentStorage, Hasher)
    service/                        Application services (use case orchestration)
  infrastructure/
    sqlite/                         SQLite repository implementations
    contentstore/                   Disk-based content-addressable storage
    hasher/                         mrCloud hash algorithm implementation
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

OAuth2 password grant flow. The server acts as both authorization server and resource server.

### Token Lifecycle

1. Client sends `POST /token` with form data: `client_id=cloud-win`, `grant_type=password`, `username=<email>`, `password=<password>`
2. Server returns `access_token`, `refresh_token`, `expires_in` (86400 seconds = 24 hours)
3. All API calls include `?access_token=<token>` as a query parameter
4. Tokens are 64-character random hex strings generated via `crypto/rand`
5. Expired tokens are rejected with status 403

### Admin Privileges

Admin status (`is_admin`) is a per-user flag. Admin endpoints (`/api/v2/admin/*`) check this flag and return 403 for non-admin users. The seed user from config is always admin.

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

Four tables:

| Table      | Purpose                                                                                                           |
|------------|-------------------------------------------------------------------------------------------------------------------|
| `users`    | User accounts: id, email, password, is_admin, quota_bytes, created                                                |
| `nodes`    | Virtual filesystem: id, user_id, parent_id, name, home (full path), node_type, size, hash, mtime, rev, grev, tree |
| `contents` | Content registry: hash, size, ref_count, created                                                                  |
| `tokens`   | Auth tokens: id, user_id, access_token, refresh_token, csrf_token, expires_at                                     |

Schema is created automatically. Migrations run at startup if needed.

## API Endpoints

All authenticated endpoints require `?access_token=<token>`.

### Service Discovery

| Method | Path | Auth | Description                                    |
|--------|------|------|------------------------------------------------|
| GET    | `/`  | No   | Returns endpoint URLs for client configuration |

### Authentication

| Method | Path                  | Auth | Description                                                                     |
|--------|-----------------------|------|---------------------------------------------------------------------------------|
| POST   | `/token`              | No   | OAuth2 password grant (form: `client_id`, `grant_type`, `username`, `password`) |
| GET    | `/api/v2/tokens/csrf` | Yes  | Get CSRF token for current session                                              |

### Dispatcher

| Method | Path                  | Auth | Description                           |
|--------|-----------------------|------|---------------------------------------|
| POST   | `/api/v2/dispatcher/` | Yes  | Returns shard URLs for all operations |
| GET    | `/d`                  | Yes  | Download shard URL (plain text)       |
| GET    | `/u`                  | Yes  | Upload shard URL (plain text)         |

### File Operations

| Method | Path                  | Auth | Parameters                         | Description              |
|--------|-----------------------|------|------------------------------------|--------------------------|
| GET    | `/api/v2/folder`      | Yes  | `home`, `offset`, `limit`, `sort`  | List directory contents  |
| GET    | `/api/v2/file`        | Yes  | `home`                             | Get file/folder metadata |
| POST   | `/api/v2/folder/add`  | Yes  | `home`, `conflict`                 | Create folder            |
| POST   | `/api/v2/file/add`    | Yes  | `home`, `hash`, `size`, `conflict` | Register uploaded file   |
| POST   | `/api/v2/file/remove` | Yes  | `home`                             | Delete file or folder    |
| POST   | `/api/v2/file/rename` | Yes  | `home`, `name`                     | Rename file or folder    |
| POST   | `/api/v2/file/move`   | Yes  | `home`, `folder`, `conflict`       | Move file or folder      |
| POST   | `/api/v2/file/copy`   | Yes  | `home`, `folder`, `conflict`       | Copy file or folder      |

### Upload / Download

| Method | Path          | Auth | Description                              |
|--------|---------------|------|------------------------------------------|
| PUT    | `/upload/`    | Yes  | Upload file binary, returns 40-char hash |
| GET    | `/get/<path>` | Yes  | Download file binary                     |

### Quota

| Method | Path                 | Auth | Description                                      |
|--------|----------------------|------|--------------------------------------------------|
| GET    | `/api/v2/user/space` | Yes  | Returns `bytes_total`, `bytes_used`, `overquota` |

### Admin User Management

All admin endpoints require the authenticated user to have `is_admin = true`.

| Method | Path                        | Parameters                                                 | Description                       |
|--------|-----------------------------|------------------------------------------------------------|-----------------------------------|
| POST   | `/api/v2/admin/user/add`    | `email`, `password`, `is_admin` (0/1), `quota_bytes`       | Create user                       |
| GET    | `/api/v2/admin/user/list`   | --                                                         | List all users with disk usage    |
| POST   | `/api/v2/admin/user/edit`   | `id`, `email`, `password`, `is_admin` (0/1), `quota_bytes` | Update user                       |
| POST   | `/api/v2/admin/user/remove` | `id`                                                       | Delete user (self-delete blocked) |

### Admin Panel

| Method | Path     | Description                          |
|--------|----------|--------------------------------------|
| GET    | `/admin` | Web-based admin panel (embedded SPA) |

## API Response Format

All API v2 responses use a standard envelope:

```json
{
  "email": "user@example.com",
  "body": { ... },
  "time": 1700490243535,
  "status": 200
}
```

- `status: 200` indicates success; the payload is in `body`
- Error responses encode the error in `body` (e.g., `body.home.error` or a plain string)
- `time` is server time in milliseconds

### Common Error Codes

| Status | Body            | Meaning                         |
|--------|-----------------|---------------------------------|
| 400    | `"exists"`      | Resource already exists         |
| 400    | `"required"`    | Missing required field          |
| 400    | `"invalid"`     | Invalid input                   |
| 400    | `"self_delete"` | Admin cannot delete own account |
| 403    | `"user"`        | Invalid or expired token        |
| 403    | `"forbidden"`   | Insufficient privileges         |
| 404    | `"not_found"`   | Resource not found              |
| 507    | `"overquota"`   | Storage quota exceeded          |

## Conflict Modes

Used in file/folder operations via the `conflict` parameter:

| Value     | Behavior                       |
|-----------|--------------------------------|
| `strict`  | Error if target already exists |
| `rename`  | Auto-rename to avoid conflict  |
| `replace` | Overwrite existing target      |

## Quota Management

- Each user has a `quota_bytes` limit
- Usage is calculated as the sum of all file node sizes for the user
- Uploads are blocked when `used + file_size > quota`
- Default quota for new users comes from `storage.quota_bytes` in config
- Quota can be set per-user via the admin API (`quota_bytes` parameter on add/edit)
- Current usage is visible in the admin panel and via `GET /api/v2/user/space`

## Startup Sequence

1. Parse `-config` flag (default: `config.yaml`)
2. Load and validate configuration
3. Open SQLite database (auto-create schema and run migrations)
4. Initialize content storage directory
5. Seed admin user from config (upsert by email)
6. Create root node (`/`) for seed user if missing
7. Register HTTP routes
8. Listen on configured address

## Dependencies

| Package              | Purpose                                 |
|----------------------|-----------------------------------------|
| `gopkg.in/yaml.v3`   | YAML configuration parsing              |
| `modernc.org/sqlite` | Pure-Go SQLite driver (no CGO required) |
| Standard library     | Everything else                         |

# Tucha Cloud API v2 -- Server Specification

Server-side protocol specification for the Tucha cloud API v2. This document describes endpoints, request/response formats, authentication, error handling, upload/download mechanics, hashing, and data models as seen from the server perspective.

**API Version**: v2
**Base URL**: `<server_url>/api/v2`

---

## Table of Contents

1. [Authentication](#1-authentication)
2. [Generic API Response Envelope](#2-generic-api-response-envelope)
3. [Shard Resolution](#3-shard-resolution)
4. [Directory Listing](#4-directory-listing)
5. [File/Folder Information](#5-filefolder-information)
6. [File Operations](#6-file-operations)
7. [Upload Protocol](#7-upload-protocol)
8. [Download Protocol](#8-download-protocol)
9. [Cloud Hash Algorithm](#9-cloud-hash-algorithm)
10. [Sharing and Publishing](#10-sharing-and-publishing)
11. [Trashbin](#11-trashbin)
12. [Thumbnails](#12-thumbnails)
13. [Video Streaming](#13-video-streaming)
14. [URL Encoding Rules](#14-url-encoding-rules)
15. [Error Codes](#15-error-codes)
16. [Data Models](#16-data-models)
17. [Constraints and Limits](#17-constraints-and-limits)
18. [Complete Endpoint Reference](#18-complete-endpoint-reference)

---

## 1. Authentication

### 1.1 OAuth2 Password Grant

**Endpoint:** `POST <server_url>/token`
**Content-Type:** `application/x-www-form-urlencoded`

**Request body:**

```
client_id=cloud-win&grant_type=password&username=<user>@<domain>&password=<url_encoded_password>
```

| Parameter    | Description                        |
|--------------|------------------------------------|
| `client_id`  | Always `cloud-win`                 |
| `grant_type` | Always `password`                  |
| `username`   | Full email address (`user@domain`) |
| `password`   | URL-encoded app password           |

**Success response (HTTP 200):**

```json
{
  "expires_in": 86400,
  "refresh_token": "<refresh_token>",
  "access_token": "<access_token>",
  "error": "",
  "error_code": 0,
  "error_description": ""
}
```

**Error response:**

```json
{
  "error": "<error_code>",
  "error_code": 123,
  "error_description": "Human-readable error"
}
```

Authentication succeeds when `error_code == 0`, fails when `error_code != 0`.

### 1.2 CSRF Token

**Endpoint:** `GET <server_url>/api/v2/tokens/csrf`

**Response:**

```json
{
  "body": {
    "token": "<csrf_token>"
  },
  "status": 200
}
```

The token is at JSON path `body.token`.

### 1.3 Request Authentication

There are two distinct authentication mechanisms depending on the endpoint type.

#### API v2 Endpoints

All API v2 requests (both GET and POST) append authentication as a query parameter:

```
?access_token=<access_token>
```

For POST requests, authentication goes in the URL query string, not the body.

**Example:**

```
POST <server_url>/api/v2/file/move?access_token=abc123
Content-Type: application/x-www-form-urlencoded

home=/path/to/file&folder=/target/folder&conflict
```

#### Upload/Download Shard Endpoints

Shard endpoints (upload, download, thumbnails) use a different parameter format:

```
?client_id=cloud-win&token=<access_token>
```

**Note:** The OAuth dispatcher uses only `token=` (without `client_id`):

```
<server_url>/d?token=<access_token>
```

### 1.4 X-CSRF-Token Header

Every authenticated API request carries a custom header:

```
X-CSRF-Token: <csrf_token>
```

### 1.5 Token Expiration Indicators

The server signals token expiration through several response patterns:

| Indicator       | Response Pattern                   |
|-----------------|------------------------------------|
| CSRF expired    | `{"body": "token", "status": 200}` |
| Session expired | `{"error": "NOT/AUTHORIZED"}`      |
| Forbidden       | `{"status": 403, "body": "user"}`  |

---

## 2. Generic API Response Envelope

All API v2 responses share this structure:

```json
{
  "email": "user@domain.ru",
  "body": { ... },
  "time": 1700490243535,
  "status": 200
}
```

| Field    | Type            | Description                                       |
|----------|-----------------|---------------------------------------------------|
| `email`  | string          | Current user email                                |
| `body`   | object/string   | Response payload (or error indicator)             |
| `time`   | integer         | Server timestamp in milliseconds                  |
| `status` | integer         | HTTP-like status code (200 = success)             |

### 2.1 Success

`status == 200` means success. The `body` field contains the response payload.

### 2.2 Errors

When `status != 200`, errors are encoded in one of three locations within `body`, tried in this order:

1. `body.home.error` (string)
2. `body.weblink.error` (string)
3. `body.invite_email.error` (string)

```json
{
  "body": {
    "home": {
      "error": "exists"
    }
  },
  "status": 400
}
```

Special HTTP status codes are mapped directly without checking the error string:
- `451` -- content blocked by rights holder
- `507` -- over quota
- `406` -- cannot add this user

### 2.3 Registration Response (file/add)

The `file/add` endpoint uses a simpler response interpretation:

- `status == 200`: Hash exists in cloud storage, file created (deduplication success)
- `status == 400`: Hash not found in cloud storage (not a real error -- upload required)

---

## 3. Shard Resolution

Shards are server endpoints for specific operations (download, upload, thumbnails, etc.). The server provides shard URLs through dispatcher endpoints.

### 3.1 API Dispatcher

**Endpoint:** `POST <server_url>/api/v2/dispatcher/`
**Authentication:** `?access_token=<token>`
**Body:** Empty string

**Response:**

```json
{
  "body": {
    "get": [{"url": "<shard_url>/get/"}],
    "upload": [{"url": "<upload_shard_url>/upload/"}],
    "thumbnails": [{"url": "<thumbnail_shard_url>/thumb/"}],
    "weblink_get": [{"url": "<shard_url>/weblink/"}],
    "weblink_view": [{"url": "..."}],
    "weblink_video": [{"url": "..."}],
    "weblink_thumbnails": [{"url": "..."}],
    "video": [{"url": "..."}],
    "view_direct": [{"url": "..."}],
    "stock": [{"url": "..."}],
    "public_upload": [{"url": "..."}],
    "auth": [{"url": "..."}],
    "web": [{"url": "..."}]
  },
  "status": 200
}
```

Each shard type is an array of objects with a `url` field.

### 3.2 OAuth Dispatcher

**Endpoint:** `GET <server_url>/<suffix>?token=<access_token>`

| Suffix | Purpose         |
|--------|-----------------|
| `d`    | Download shard  |
| `u`    | Upload shard    |

**Response:** Plain text, format: `URL IP COUNT` (space-separated).

Example: `<shard_url>/get/ 1.2.3.4 5`

The first space-delimited word is the shard URL.

### 3.3 Shard Types

| String Value          | Description                        |
|-----------------------|------------------------------------|
| `get`                 | Download files                     |
| `upload`              | Upload files                       |
| `thumbnails`          | Thumbnail images                   |
| `weblink_get`         | Download public files              |
| `weblink_view`        | View public files                  |
| `weblink_video`       | Stream public videos               |
| `weblink_thumbnails`  | Public file thumbnails             |
| `video`               | Video streaming                    |
| `view_direct`         | Direct view (internal)             |
| `stock`               | Stock files                        |
| `public_upload`       | Public upload                      |
| `auth`                | Authentication                     |
| `web`                 | Web interface                      |

---

## 4. Directory Listing

### 4.1 List Directory

**Endpoint:** `GET <server_url>/api/v2/folder`

**Query parameters for regular accounts:**

```
sort={"type":"name","order":"asc"}&offset=<offset>&limit=<limit>&home=<url_encoded_path>&access_token=<token>
```

**Query parameters for public (shared) accounts:**

```
sort={"type":"name","order":"asc"}&offset=0&limit=<limit>&weblink=<public_link_with_trailing_slash><url_encoded_path>&access_token=<token>
```

| Parameter | Description                                       |
|-----------|---------------------------------------------------|
| `sort`    | URL-encoded JSON: `{"type":"name","order":"asc"}` |
| `offset`  | Pagination offset (0-based)                       |
| `limit`   | Max items per page (up to 65535)                  |
| `home`    | URL-encoded cloud path for regular accounts       |
| `weblink` | Public link identifier + path for public accounts |

**Response:**

```json
{
  "body": {
    "count": {"folders": 1, "files": 1},
    "tree": "316234396237373230303030",
    "name": "TEST_DIR",
    "grev": 13501,
    "size": 44629,
    "sort": {"order": "asc", "type": "name"},
    "kind": "folder",
    "rev": 13499,
    "type": "folder",
    "home": "/TEST_DIR",
    "list": [
      {
        "count": {"folders": 1, "files": 0},
        "tree": "316234396237373230303030",
        "name": "subdir",
        "grev": 13501,
        "size": 668,
        "kind": "folder",
        "rev": 13500,
        "type": "folder",
        "home": "/TEST_DIR/subdir"
      },
      {
        "mtime": 1700490201,
        "virus_scan": "pass",
        "name": "sign.png",
        "size": 43961,
        "hash": "C172C6E2FF47284FF33F348FEA7EECE532F6C051",
        "kind": "file",
        "type": "file",
        "home": "/TEST_DIR/sign.png"
      }
    ]
  },
  "status": 200
}
```

### 4.2 Directory Item Fields

**Common fields (all items):**

| JSON Key  | Type    | Description                                          |
|-----------|---------|------------------------------------------------------|
| `name`    | string  | Item name                                            |
| `type`    | string  | `"folder"` or `"file"`                               |
| `kind`    | string  | `"folder"`, `"file"`, or `"shared"` (mounted shares) |
| `home`    | string  | Full cloud path                                      |
| `size`    | int64   | Size in bytes                                        |
| `weblink` | string  | Public link identifier (present if published)        |
| `grev`    | integer | Global revision number                               |
| `rev`     | integer | Revision number                                      |

**File-specific fields (`type == "file"`):**

| JSON Key     | Type   | Description                                 |
|--------------|--------|---------------------------------------------|
| `mtime`      | int64  | Modification time (Unix timestamp, seconds) |
| `hash`       | string | Cloud hash (40-char uppercase hex)          |
| `virus_scan` | string | Virus scan status (e.g., `"pass"`)          |

**Folder-specific fields (`type == "folder"`):**

| JSON Key        | Type    | Description               |
|-----------------|---------|---------------------------|
| `tree`          | string  | Tree identifier (hex)     |
| `count.folders` | integer | Number of subfolders      |
| `count.files`   | integer | Number of files           |

**Trashbin item fields (additional):**

| JSON Key        | Type    | Description                               |
|-----------------|---------|-------------------------------------------|
| `deleted_at`    | integer | Unix timestamp of deletion                |
| `deleted_from`  | string  | Original path before deletion             |
| `deleted_by`    | integer | Who deleted                               |

### 4.3 Pagination

The API silently caps responses at approximately 8000 items per request regardless of the `limit` parameter. Pagination uses the `offset` parameter:

1. First request: `offset=0`
2. Response includes `body.count.files + body.count.folders` as the total expected count
3. Subsequent requests: `offset=<total_items_received_so_far>`
4. Continue until all items received or an empty page is returned

---

## 5. File/Folder Information

### 5.1 File Status

**Endpoint:** `GET <server_url>/api/v2/file`

**Regular accounts:**

```
GET <server_url>/api/v2/file?home=<url_encoded_path>&access_token=<token>
```

**Public accounts:**

```
GET <server_url>/api/v2/file?weblink=<public_link_with_slash><url_encoded_path>&access_token=<token>
```

**Response:** Standard envelope with `body` containing a single directory item (same fields as listing items).

### 5.2 User Space

**Endpoint:** `GET <server_url>/api/v2/user/space?access_token=<token>`

**Response:**

```json
{
  "body": {
    "overquota": false,
    "bytes_total": 17179869184,
    "bytes_used": 5368709120
  },
  "status": 200
}
```

---

## 6. File Operations

All file operations use `POST` with `application/x-www-form-urlencoded` content type. Authentication is via `?access_token=<token>` query parameter appended to the URL.

### 6.1 Create Folder

**Endpoint:** `POST <server_url>/api/v2/folder/add?access_token=<token>`

**Body:** `home=/<url_encoded_path>&conflict`

Note: `&conflict` appears as a bare parameter with no value.

### 6.2 Delete File

**Endpoint:** `POST <server_url>/api/v2/file/remove?access_token=<token>`

**Body:** `home=/<url_encoded_path>&conflict`

### 6.3 Remove Directory

**Endpoint:** `POST <server_url>/api/v2/file/remove?access_token=<token>`

Same endpoint as file deletion but with trailing slash on the path.

**Body:** `home=/<url_encoded_path_with_trailing_slash>&conflict`

Note: The API reportedly always returns success even if the path does not exist.

### 6.4 Rename File/Folder

**Endpoint:** `POST <server_url>/api/v2/file/rename?access_token=<token>`

**Body:** `home=<url_encoded_current_path>&name=<url_encoded_new_name>`

Note: `home` does NOT have a leading `/` prepended (unlike create/delete). `name` is just the new name, not the full path.

### 6.5 Move File/Folder

**Endpoint:** `POST <server_url>/api/v2/file/move?access_token=<token>`

**Body:** `home=<url_encoded_source_path>&folder=<url_encoded_target_folder>&conflict`

Note: No leading `/` prefix on either parameter.

### 6.6 Copy File/Folder

**Endpoint:** `POST <server_url>/api/v2/file/copy?access_token=<token>`

**Body:** `home=/<url_encoded_source_path>&folder=/<url_encoded_target_folder>&conflict`

Note: Both `home` and `folder` have explicit leading `/` prefix (unlike Move).

### 6.7 Conflict Modes

The `conflict` parameter controls behavior when the target already exists:

| Value    | Description                     |
|----------|---------------------------------|
| `strict` | Return error if exists          |
| `rename` | Auto-rename the new item        |
| `ignore` | Apparently not implemented      |

---

## 7. Upload Protocol

Upload is a two-step process: binary upload to a shard, then registration via API.

### 7.1 Step 1: PUT File to Upload Shard

**Method:** `PUT`
**URL:** `<upload_shard>?client_id=cloud-win&token=<access_token>`
**Body:** Raw file stream (binary, sent directly as PUT body -- no multipart encoding)

**Success response (HTTP 200):** Plain text containing the 40-character uppercase hex hash of the uploaded file content.

Example response body: `C172C6E2FF47284FF33F348FEA7EECE532F6C051`

If the response length is not exactly 40 characters, the upload has failed.

### 7.2 Step 2: Register File by Hash

**Endpoint:** `POST <server_url>/api/v2/file/add?access_token=<token>`
**Content-Type:** `application/x-www-form-urlencoded`

**Body:** `api=2&conflict=<conflict_mode>&home=/<url_encoded_remote_path>&hash=<40_char_hash>&size=<file_size_bytes>`

| Parameter  | Description                            |
|------------|----------------------------------------|
| `api`      | Always `2`                             |
| `conflict` | Conflict mode (see Section 6.7)        |
| `home`     | Target cloud path with leading `/`     |
| `hash`     | 40-char uppercase hex hash from upload |
| `size`     | File size in bytes (decimal integer)   |

**Response:**
- `status == 200`: File created successfully
- `status == 400`: Hash not found in cloud storage (see Section 7.3)

### 7.3 Deduplication (Fast Upload)

The `file/add` endpoint (Step 2) can be called without performing Step 1. If the server already has content with the given hash, the file is created instantly without uploading:

- `status == 200`: Hash exists -- file created (no upload needed)
- `status == 400`: Hash not found -- content must be uploaded first via Step 1

This allows content-addressable storage: identical files are stored once regardless of how many paths reference them.

---

## 8. Download Protocol

### 8.1 Regular Account Download

**Method:** `GET`
**URL:** `<download_shard><url_encoded_path>?client_id=cloud-win&token=<access_token>`

**Required headers:**
- `User-Agent: cloud-win` -- the download endpoint **blocks browser-like User-Agents** (e.g., `Mozilla/*`)

**Response:** Raw file stream (binary).

### 8.2 Public Account Download

**URL:** `<public_shard>/<public_link>/<url_encoded_path>`

No authentication parameters needed for public downloads.

### 8.3 Public Shard Extraction

For public accounts, the download shard is extracted from the public URL page HTML. The page contains an embedded JSON fragment:

```
"weblink_get": { "url": "https://..." }
```

The `url` value within the `weblink_get` block is the public download shard.

---

## 9. Cloud Hash Algorithm

The cloud uses a proprietary SHA1-based hash algorithm for file identification and deduplication.

### 9.1 Small Files (< 21 bytes)

```
1. Allocate a 20-byte buffer, zero-initialized
2. Read file content into the buffer (remaining bytes stay zero)
3. Return UpperCase(hex_string_of_raw_20_bytes)
```

This is NOT a SHA1 hash -- it is the raw file content zero-padded to 20 bytes, represented as a 40-character hex string.

### 9.2 Large Files (>= 21 bytes)

```
SHA1("mrCloud" + file_content + decimal_file_size_as_string)
```

Detailed steps:
1. Initialize SHA1 context
2. Feed seed bytes: UTF-8 encoding of `"mrCloud"` (7 bytes: `6D 72 43 6C 6F 75 64`)
3. Feed entire file content
4. Feed size suffix: UTF-8 encoding of the file size as a decimal string (e.g., `"12345"`)
5. Finalize SHA1, return uppercase 40-character hex string

### 9.3 Constants

| Constant             | Value         |
|----------------------|---------------|
| Hash seed            | `"mrCloud"`   |
| Small file threshold | 21 bytes      |
| Small file buffer    | 20 bytes      |
| Hash hex length      | 40 characters |

---

## 10. Sharing and Publishing

### 10.1 Publish File/Folder (Create Public Link)

**Endpoint:** `POST <server_url>/api/v2/file/publish?access_token=<token>`
**Content-Type:** `application/x-www-form-urlencoded`

**Body:** `home=/<url_encoded_path>&conflict`

**Response:**

```json
{
  "body": "<public_link_id>",
  "status": 200
}
```

The `body` field is a plain string (not an object) containing the public link identifier. The public URL is: `<server_url>/public/<public_link_id>`

### 10.2 Unpublish

**Endpoint:** `POST <server_url>/api/v2/file/unpublish?access_token=<token>`
**Content-Type:** `application/x-www-form-urlencoded`

**Body:** `weblink=<public_link_id>&conflict`

### 10.3 Share Folder

**Endpoint:** `POST <server_url>/api/v2/folder/share?access_token=<token>`

**Body:** `home=/<url_encoded_path>&invite={"email":"<email>","access":"<access>"}`

Where `<access>` is:
- `read_only`
- `read_write`

### 10.4 Unshare Folder

**Endpoint:** `POST <server_url>/api/v2/folder/unshare?access_token=<token>`

**Body:** `home=/<url_encoded_path>&invite={"email":"<email>"}`

### 10.5 Get Share Info

**Endpoint:** `GET <server_url>/api/v2/folder/shared/info?home=<url_encoded_path>&access_token=<token>`

**Response:**

```json
{
  "body": {
    "invited": [
      {
        "email": "user@example.com",
        "status": "accepted",
        "access": "read_write",
        "name": "User Name"
      }
    ]
  },
  "status": 200
}
```

If `body.invited` is null or not an array, the folder has no invites.

### 10.6 Get Shared Links

**Endpoint:** `GET <server_url>/api/v2/folder/shared/links?access_token=<token>`

**Response:** Standard envelope with `body.list` array of directory items that have `weblink` field set.

### 10.7 Get Incoming Invites

**Endpoint:** `GET <server_url>/api/v2/folder/shared/incoming?access_token=<token>`

**Response:**

```json
{
  "body": {
    "list": [
      {
        "owner": {"email": "owner@example.com", "name": "Owner"},
        "tree": "...",
        "access": "read_write",
        "name": "SharedFolder",
        "home": "/SharedFolder",
        "size": 12345,
        "invite_token": "<token>"
      }
    ]
  },
  "status": 200
}
```

The `home` field is only present for already-mounted invites.

### 10.8 Mount Shared Folder

**Endpoint:** `POST <server_url>/api/v2/folder/mount?access_token=<token>`

**Body:** `home=<url_encoded_name>&invite_token=<token>&conflict=<mode>`

Default conflict mode: `rename`

### 10.9 Unmount Shared Folder

**Endpoint:** `POST <server_url>/api/v2/folder/unmount?access_token=<token>`

**Body:** `home=<url_encoded_path>&clone_copy=<true|false>`

When `clone_copy=true`, a copy of the shared content is kept.

### 10.10 Reject Invite

**Endpoint:** `POST <server_url>/api/v2/folder/invites/reject?access_token=<token>`

**Body:** `invite_token=<token>`

### 10.11 Clone Public Link

**Endpoint:** `GET <server_url>/api/v2/clone`

**Query parameters:** `folder=/<url_encoded_target_path>&weblink=<link_id>&conflict=<mode>&access_token=<token>`

Default conflict mode: `rename`

---

## 11. Trashbin

### 11.1 List Trashbin

**Endpoint:** `GET <server_url>/api/v2/trashbin?access_token=<token>`

**Response:** Standard envelope with `body.list` array containing trashbin items (directory items with additional `deleted_at`, `deleted_from`, `deleted_by` fields).

### 11.2 Restore from Trashbin

**Endpoint:** `POST <server_url>/api/v2/trashbin/restore?access_token=<token>`

**Body:** `path=<url_encoded_original_path>&restore_revision=<rev>&conflict=<mode>`

| Parameter          | Description                                               |
|--------------------|-----------------------------------------------------------|
| `path`             | Full original path (`deleted_from` + `name`), URL-encoded |
| `restore_revision` | The `rev` field from the trash item (integer)             |
| `conflict`         | Default: `rename`                                         |

### 11.3 Empty Trashbin

**Endpoint:** `POST <server_url>/api/v2/trashbin/empty?access_token=<token>`

**Body:** Empty string

---

## 12. Thumbnails

### 12.1 Thumbnail URL Format

```
<thumbnail_shard>/<size_preset>/<url_encoded_cloud_path>?client_id=cloud-win&token=<access_token>
```

Example: `<thumbnail_shard_url>/thumb/xw14/folder/image.jpg?client_id=cloud-win&token=abc123`

**Fallback shard:** `<thumbnail_shard_url>/thumb`

### 12.2 Size Presets

| Preset | Width    | Height   |
|--------|----------|----------|
| `xw11` | 26       | 26       |
| `xw27` | 28       | 38       |
| `xw22` | 36       | 24       |
| `xw12` | 52       | 35       |
| `xw28` | 64       | 43       |
| `xw23` | 72       | 48       |
| `xw14` | 160      | 107      |
| `xw17` | 160      | 120      |
| `xw10` | 160      | 120      |
| `xw29` | 150      | 150      |
| `xw24` | 168      | 112      |
| `xw20` | 170      | 113      |
| `xw15` | 206      | 137      |
| `xw19` | 206      | 206      |
| `xw26` | 270      | 365      |
| `xw18` | 305      | 230      |
| `xw13` | 320      | 213      |
| `xw16` | 320      | 240      |
| `xw25` | 336      | 224      |
| `xw21` | 340      | 226      |
| `xw2`  | 1000     | 667      |
| `xw1`  | original | original |

The server handles aspect ratio scaling within the requested preset dimensions.

---

## 13. Video Streaming

### 13.1 HLS Streaming URL

For video files with public weblinks:

```
<weblink_video_shard>0p/<base64_encoded_weblink>.m3u8?double_encode=1
```

The weblink string is Base64-encoded (standard encoding, treating the weblink as raw bytes).

Default shard type: `weblink_video`.

If the file has no weblink, it must be published first (via `file/publish`).

### 13.2 Non-Playlist Streaming URL

For other streaming formats, the URL follows the same pattern as public downloads:

```
<shard>/<public_link>/<url_encoded_path>
```

The shard type varies by streaming mode (e.g., `weblink_view`, `video`, `view_direct`, `thumbnails`, `weblink_thumbnails`).

---

## 14. URL Encoding Rules

### 14.1 URL Encoding

The server expects URL parameters encoded as follows:

**Characters NOT encoded (pass-through):**

```
a-z  A-Z  0-9  /  _  -  .
```

**All other characters** are encoded as `%XX` (two uppercase hex digits) after UTF-8 conversion.

Note: Forward slashes `/` are preserved. Spaces become `%20`. Cyrillic and other non-ASCII characters are multi-byte `%XX%XX` sequences.

### 14.2 Path Conventions

- Cloud paths use forward slashes (`/`) as separators
- POST body paths are URL-encoded
- Download shard URL is suffixed with the URL-encoded file path
- Upload shard URL has no path suffix -- the file is sent as the PUT body
- Public download URL: `<shard>/<public_link>/<url_encoded_path>`
- An empty path is typically represented as `'/'`
- Trailing slashes on directory paths are significant for `file/remove` (directories require trailing `/`)

---

## 15. Error Codes

### 15.1 API Error Strings

Error strings returned in `body.home.error`, `body.weblink.error`, or `body.invite_email.error`:

| Error String            | Description                                          |
|-------------------------|------------------------------------------------------|
| `exists`                | Item already exists                                  |
| `required`              | Name cannot be empty                                 |
| `invalid`               | Invalid name (forbidden characters)                  |
| `readonly`              | Read-only access                                     |
| `read_only`             | Read-only access (alternative form)                  |
| `name_length_exceeded`  | Name exceeds server limit                            |
| `overquota`             | Not enough cloud space                               |
| `quota_exceeded`        | Not enough cloud space (alternative form)            |
| `not_exists`            | Item does not exist                                  |
| `own`                   | Cannot clone own link                                |
| `name_too_long`         | File name too long                                   |
| `virus_scan_fail`       | File is infected                                     |
| `owner`                 | Cannot use own email for sharing                     |
| `trees_conflict`        | Cannot share folder containing/inside shared folders |
| `unprocessable_entry`   | Cannot grant file access                             |
| `user_limit_exceeded`   | Max users per shared folder exceeded                 |
| `export_limit_exceeded` | Max shared folders per account exceeded              |
| `unknown`               | Unspecified server error                             |

### 15.2 HTTP Status Codes

| Status | Meaning                                              |
|--------|------------------------------------------------------|
| 200    | Success                                              |
| 400    | Bad request (also: hash not found for deduplication) |
| 403    | Token expired / session invalid                      |
| 406    | Cannot add this user                                 |
| 451    | Content blocked by rights holder                     |
| 500    | Item exists (server-side)                            |
| 507    | Over quota                                           |

---

## 16. Data Models

JSON structures returned by the server.

### 16.1 Directory Item (in `body.list[]` or single-item `body`)

```json
{
  "tree": "316234396237373230303030",
  "name": "example.txt",
  "grev": 13501,
  "size": 44629,
  "kind": "file",
  "weblink": "abc123/def456",
  "rev": 13499,
  "type": "file",
  "home": "/path/example.txt",
  "mtime": 1700490201,
  "hash": "C172C6E2FF47284FF33F348FEA7EECE532F6C051",
  "virus_scan": "pass"
}
```

See Section 4.2 for complete field descriptions.

### 16.2 OAuth Token

```json
{
  "error": "",
  "error_code": 0,
  "error_description": "",
  "expires_in": 86400,
  "refresh_token": "<refresh_token>",
  "access_token": "<access_token>"
}
```

### 16.3 Storage Quota

```json
{
  "overquota": false,
  "bytes_total": 17179869184,
  "bytes_used": 5368709120
}
```

### 16.4 File Identity (used in `file/add` registration)

A file is uniquely identified by:

| Field  | Description                     |
|--------|---------------------------------|
| `hash` | 40-char uppercase hex hash      |
| `size` | File size in bytes (int64)      |

### 16.5 Incoming Invite

```json
{
  "owner": {"email": "owner@example.com", "name": "Owner"},
  "tree": "...",
  "access": "read_write",
  "name": "SharedFolder",
  "home": "/SharedFolder",
  "size": 12345,
  "invite_token": "<token>"
}
```

The `home` field is empty/absent when the invite has not been mounted yet.

### 16.6 Share Member (Invite)

```json
{
  "email": "user@example.com",
  "status": "accepted",
  "access": "read_write",
  "name": "User Name"
}
```

### 16.7 Access Levels

| API String    | Description      |
|---------------|------------------|
| `read_only`   | Read-only access |
| `read_write`  | Read and write   |

---

## 17. Constraints and Limits

| Constraint                         | Value                               |
|------------------------------------|-------------------------------------|
| Max file name length               | 255 characters                      |
| Listing page size (max `limit`)    | 65,535                              |
| Actual items per page (server cap) | ~8,000                              |
| Hash hex length                    | 40 characters                       |
| Small file hash threshold          | 21 bytes                            |
| Max users per shared folder        | 200                                 |
| Max shared folders per account     | 50                                  |
| OAuth `client_id`                  | `cloud-win`                         |
| Hash seed                          | `mrCloud` (7 UTF-8 bytes)           |
| Public URL prefix                  | `<server_url>/public/`     |
| Thumbnail fallback URL             | `<thumbnail_shard_url>/thumb` |

---

## 18. Complete Endpoint Reference

| #  | Method | Endpoint                                         | Purpose                      |
|----|--------|--------------------------------------------------|------------------------------|
| 1  | POST   | `<server_url>/token`                       | OAuth authentication         |
| 2  | GET    | `/api/v2/tokens/csrf`                            | Get CSRF token               |
| 3  | POST   | `/api/v2/dispatcher/`                            | Resolve shard URLs           |
| 4  | GET    | `<server_url>/<d\|u>`        | OAuth dispatcher (plaintext) |
| 5  | GET    | `/api/v2/folder`                                 | List directory               |
| 6  | GET    | `/api/v2/file`                                   | Get file/folder info         |
| 7  | POST   | `/api/v2/folder/add`                             | Create folder                |
| 8  | POST   | `/api/v2/file/remove`                            | Delete file/folder           |
| 9  | POST   | `/api/v2/file/rename`                            | Rename file/folder           |
| 10 | POST   | `/api/v2/file/move`                              | Move file/folder             |
| 11 | POST   | `/api/v2/file/copy`                              | Copy file/folder             |
| 12 | POST   | `/api/v2/file/add`                               | Register file by hash        |
| 13 | PUT    | `<upload_shard>?client_id=...&token=...`         | Upload file binary           |
| 14 | GET    | `<download_shard><path>?client_id=...&token=...` | Download file binary         |
| 15 | POST   | `/api/v2/file/publish`                           | Create public link           |
| 16 | POST   | `/api/v2/file/unpublish`                         | Remove public link           |
| 17 | POST   | `/api/v2/folder/share`                           | Share folder with user       |
| 18 | POST   | `/api/v2/folder/unshare`                         | Revoke folder sharing        |
| 19 | GET    | `/api/v2/folder/shared/info`                     | Get share members            |
| 20 | GET    | `/api/v2/folder/shared/links`                    | List published items         |
| 21 | GET    | `/api/v2/folder/shared/incoming`                 | List incoming invites        |
| 22 | POST   | `/api/v2/folder/mount`                           | Accept & mount invite        |
| 23 | POST   | `/api/v2/folder/unmount`                         | Unmount shared folder        |
| 24 | POST   | `/api/v2/folder/invites/reject`                  | Reject invite                |
| 25 | GET    | `/api/v2/clone`                                  | Clone public link            |
| 26 | GET    | `/api/v2/user/space`                             | Get storage quota            |
| 27 | GET    | `/api/v2/trashbin`                               | List trashbin                |
| 28 | POST   | `/api/v2/trashbin/restore`                       | Restore from trash           |
| 29 | POST   | `/api/v2/trashbin/empty`                         | Empty trashbin               |
| 30 | GET    | `<thumbnail_shard>/<preset>/<path>?...`          | Download thumbnail           |

All `/api/v2/*` endpoints are relative to `<server_url>`.

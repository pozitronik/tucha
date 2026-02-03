// Package sqlite provides SQLite-based implementations of domain repository interfaces.
package sqlite

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

const schema = `
CREATE TABLE IF NOT EXISTS users (
    id        INTEGER PRIMARY KEY AUTOINCREMENT,
    email     TEXT NOT NULL UNIQUE,
    password  TEXT NOT NULL,
    created   INTEGER NOT NULL DEFAULT (strftime('%s','now'))
);

CREATE TABLE IF NOT EXISTS nodes (
    id        INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id   INTEGER NOT NULL REFERENCES users(id),
    parent_id INTEGER REFERENCES nodes(id) ON DELETE CASCADE,
    name      TEXT NOT NULL,
    home      TEXT NOT NULL,
    node_type TEXT NOT NULL CHECK (node_type IN ('file','folder')),
    size      INTEGER NOT NULL DEFAULT 0,
    hash      TEXT,
    mtime     INTEGER NOT NULL DEFAULT (strftime('%s','now')),
    rev       INTEGER NOT NULL DEFAULT 1,
    grev      INTEGER NOT NULL DEFAULT 1,
    tree      TEXT NOT NULL DEFAULT '',
    created   INTEGER NOT NULL DEFAULT (strftime('%s','now')),
    UNIQUE(user_id, home)
);

CREATE TABLE IF NOT EXISTS contents (
    hash      TEXT PRIMARY KEY,
    size      INTEGER NOT NULL,
    ref_count INTEGER NOT NULL DEFAULT 1,
    created   INTEGER NOT NULL DEFAULT (strftime('%s','now'))
);

CREATE TABLE IF NOT EXISTS tokens (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id       INTEGER NOT NULL REFERENCES users(id),
    access_token  TEXT NOT NULL UNIQUE,
    refresh_token TEXT NOT NULL UNIQUE,
    csrf_token    TEXT NOT NULL,
    expires_at    INTEGER NOT NULL,
    created       INTEGER NOT NULL DEFAULT (strftime('%s','now'))
);
`

// DB wraps the SQLite database connection.
type DB struct {
	conn *sql.DB
}

// Open creates or opens a SQLite database at the given path, initializes
// the schema, and enables WAL mode and foreign keys.
func Open(dbPath string) (*DB, error) {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("creating database directory: %w", err)
	}

	conn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	for _, pragma := range []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA foreign_keys=ON",
	} {
		if _, err := conn.Exec(pragma); err != nil {
			conn.Close()
			return nil, fmt.Errorf("executing %s: %w", pragma, err)
		}
	}

	if _, err := conn.Exec(schema); err != nil {
		conn.Close()
		return nil, fmt.Errorf("initializing schema: %w", err)
	}

	return &DB{conn: conn}, nil
}

// Close closes the underlying database connection.
func (db *DB) Close() error {
	return db.conn.Close()
}

// Conn returns the underlying database connection for use by repository implementations.
func (db *DB) Conn() *sql.DB {
	return db.conn
}

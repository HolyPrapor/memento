package sqlite

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

func OpenDB(dbPath string) (*sql.DB, error) {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("create db directory: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping db: %w", err)
	}

	return db, nil
}

func InitSchema(db *sql.DB) error {
	statements := []string{
		`CREATE TABLE IF NOT EXISTS sections (
			id INTEGER PRIMARY KEY,
			path TEXT NOT NULL,
			anchor TEXT NOT NULL,
			heading TEXT,
			heading_level INTEGER NOT NULL,
			body TEXT NOT NULL,
			section_order INTEGER NOT NULL
		)`,
		`CREATE VIRTUAL TABLE IF NOT EXISTS sections_fts USING fts5(
			content,
			path UNINDEXED,
			anchor UNINDEXED,
			heading_level UNINDEXED,
			section_order UNINDEXED,
			tokenize='porter'
		)`,
	}

	for _, stmt := range statements {
		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("exec schema: %w", err)
		}
	}

	return nil
}

func RecreateTables(db *sql.DB) error {
	if _, err := db.Exec(`DROP TABLE IF EXISTS sections_fts`); err != nil {
		return fmt.Errorf("drop sections_fts: %w", err)
	}
	if _, err := db.Exec(`DROP TABLE IF EXISTS sections`); err != nil {
		return fmt.Errorf("drop sections: %w", err)
	}
	return InitSchema(db)
}

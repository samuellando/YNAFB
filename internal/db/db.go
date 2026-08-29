package dbutil

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "modernc.org/sqlite"
)

func Open(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite database: %w", err)
	}

	if _, err := db.Exec(`PRAGMA foreign_keys = ON;`); err != nil {
		db.Close()
		return nil, fmt.Errorf("enable foreign keys: %w", err)
	}

	if err := applyMigrations(db); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

func applyMigrations(db *sql.DB) error {
	entries, err := os.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("read migrations directory: %w", err)
	}

	filenames := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		filenames = append(filenames, entry.Name())
	}

	sort.Strings(filenames)

	for _, filename := range filenames {
		contents, err := os.ReadFile(filepath.Join("migrations", filename))
		if err != nil {
			return fmt.Errorf("read migration %s: %w", filename, err)
		}

		if _, err := db.Exec(string(contents)); err != nil {
			return fmt.Errorf("apply migration %s: %w", filename, err)
		}
	}

	return nil
}

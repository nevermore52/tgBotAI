package database

import (
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

type Migration struct {
	Version int
	Name    string
	UpSQL   string
	DownSQL string
}

func RunMigrations(db *sqlx.DB) error {
	createMigrationsTable := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`
	if _, err := db.Exec(createMigrationsTable); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	migrations, err := loadMigrations()
	if err != nil {
		return fmt.Errorf("failed to load migrations: %w", err)
	}

	appliedMigrations, err := getAppliedMigrations(db)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	for _, migration := range migrations {
		if appliedMigrations[migration.Version] {
			continue
		}

		fmt.Printf("Applying migration %d: %s\n", migration.Version, migration.Name)

		tx, err := db.Beginx()
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}

		if _, err := tx.Exec(migration.UpSQL); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to apply migration %d: %w", migration.Version, err)
		}

		insertMigration := `
			INSERT INTO schema_migrations (version, name) 
			VALUES ($1, $2)
		`
		if _, err := tx.Exec(insertMigration, migration.Version, migration.Name); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to record migration %d: %w", migration.Version, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration %d: %w", migration.Version, err)
		}

		fmt.Printf("Migration %d applied successfully\n", migration.Version)
	}

	return nil
}

func loadMigrations() ([]Migration, error) {
	var migrations []Migration
	migrationMap := make(map[int]*Migration)

	err := fs.WalkDir(migrationsFS, "migrations", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		baseName := filepath.Base(path)

		if !strings.HasSuffix(baseName, ".sql") {
			return nil
		}

		nameWithoutExt := strings.TrimSuffix(baseName, ".sql")

		var nameWithoutType string
		var isUp bool

		if strings.HasSuffix(nameWithoutExt, ".up") {
			isUp = true
			nameWithoutType = strings.TrimSuffix(nameWithoutExt, ".up")
		} else if strings.HasSuffix(nameWithoutExt, ".down") {
			nameWithoutType = strings.TrimSuffix(nameWithoutExt, ".down")
		} else {
			return nil
		}

		parts := strings.SplitN(nameWithoutType, "_", 2)
		if len(parts) < 2 {
			return nil
		}

		version, err := strconv.Atoi(parts[0])
		if err != nil {
			return nil
		}

		migrationName := parts[1]

		content, err := migrationsFS.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", path, err)
		}

		if migrationMap[version] == nil {
			migrationMap[version] = &Migration{
				Version: version,
				Name:    migrationName,
			}
		}

		if isUp {
			migrationMap[version].UpSQL = string(content)
		} else {
			migrationMap[version].DownSQL = string(content)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	for _, migration := range migrationMap {
		migrations = append(migrations, *migration)
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

func getAppliedMigrations(db *sqlx.DB) (map[int]bool, error) {
	applied := make(map[int]bool)

	rows, err := db.Query("SELECT version FROM schema_migrations")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		applied[version] = true
	}

	return applied, rows.Err()
}

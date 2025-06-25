package main

import (
	"actor-model-observability/internal/config"
	"database/sql"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"sort"
	"strings"

	_ "github.com/lib/pq"
)

func main() {
	var (
		command = flag.String("command", "up", "Migration command: up, down, status")
		steps   = flag.Int("steps", 0, "Number of migration steps (0 = all)")
	)
	flag.Parse()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Connect to database
	db, err := sql.Open("postgres", cfg.GetDSN())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// Create migrations table if it doesn't exist
	if err := createMigrationsTable(db); err != nil {
		log.Fatalf("Failed to create migrations table: %v", err)
	}

	// Execute command
	switch *command {
	case "up":
		if err := migrateUp(db, *steps); err != nil {
			log.Fatalf("Migration up failed: %v", err)
		}
	case "down":
		if err := migrateDown(db, *steps); err != nil {
			log.Fatalf("Migration down failed: %v", err)
		}
	case "status":
		if err := migrationStatus(db); err != nil {
			log.Fatalf("Migration status failed: %v", err)
		}
	default:
		log.Fatalf("Unknown command: %s", *command)
	}
}

func createMigrationsTable(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`
	_, err := db.Exec(query)
	return err
}

func migrateUp(db *sql.DB, steps int) error {
	// Get all migration files
	migrations, err := getMigrationFiles()
	if err != nil {
		return fmt.Errorf("failed to get migration files: %w", err)
	}

	// Get applied migrations
	applied, err := getAppliedMigrations(db)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Filter pending migrations
	var pending []string
	for _, migration := range migrations {
		if !contains(applied, migration) {
			pending = append(pending, migration)
		}
	}

	// Limit steps if specified
	if steps > 0 && len(pending) > steps {
		pending = pending[:steps]
	}

	if len(pending) == 0 {
		log.Println("No pending migrations")
		return nil
	}

	// Apply migrations
	for _, migration := range pending {
		log.Printf("Applying migration: %s", migration)
		if err := applyMigration(db, migration); err != nil {
			return fmt.Errorf("failed to apply migration %s: %w", migration, err)
		}
		log.Printf("Applied migration: %s", migration)
	}

	log.Printf("Successfully applied %d migrations", len(pending))
	return nil
}

func migrateDown(db *sql.DB, steps int) error {
	// Get applied migrations
	applied, err := getAppliedMigrations(db)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	if len(applied) == 0 {
		log.Println("No migrations to rollback")
		return nil
	}

	// Sort in reverse order for rollback
	sort.Sort(sort.Reverse(sort.StringSlice(applied)))

	// Limit steps if specified
	if steps > 0 && len(applied) > steps {
		applied = applied[:steps]
	}

	// Rollback migrations
	for _, migration := range applied {
		log.Printf("Rolling back migration: %s", migration)
		if err := rollbackMigration(db, migration); err != nil {
			return fmt.Errorf("failed to rollback migration %s: %w", migration, err)
		}
		log.Printf("Rolled back migration: %s", migration)
	}

	log.Printf("Successfully rolled back %d migrations", len(applied))
	return nil
}

func migrationStatus(db *sql.DB) error {
	// Get all migration files
	migrations, err := getMigrationFiles()
	if err != nil {
		return fmt.Errorf("failed to get migration files: %w", err)
	}

	// Get applied migrations
	applied, err := getAppliedMigrations(db)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	log.Println("Migration Status:")
	log.Println("=================")

	for _, migration := range migrations {
		status := "PENDING"
		if contains(applied, migration) {
			status = "APPLIED"
		}
		log.Printf("%-50s %s", migration, status)
	}

	log.Printf("\nTotal migrations: %d", len(migrations))
	log.Printf("Applied: %d", len(applied))
	log.Printf("Pending: %d", len(migrations)-len(applied))

	return nil
}

func getMigrationFiles() ([]string, error) {
	migrationsDir := "migrations"
	files, err := ioutil.ReadDir(migrationsDir)
	if err != nil {
		return nil, err
	}

	var migrations []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".sql") {
			migrations = append(migrations, file.Name())
		}
	}

	sort.Strings(migrations)
	return migrations, nil
}

func getAppliedMigrations(db *sql.DB) ([]string, error) {
	rows, err := db.Query("SELECT version FROM schema_migrations ORDER BY version")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var applied []string
	for rows.Next() {
		var version string
		if err := rows.Scan(&version); err != nil {
			return nil, err
		}
		applied = append(applied, version)
	}

	return applied, rows.Err()
}

func applyMigration(db *sql.DB, migration string) error {
	// Read migration file
	content, err := ioutil.ReadFile(filepath.Join("migrations", migration))
	if err != nil {
		return err
	}

	// Begin transaction
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Execute migration
	if _, err := tx.Exec(string(content)); err != nil {
		return err
	}

	// Record migration
	if _, err := tx.Exec("INSERT INTO schema_migrations (version) VALUES ($1)", migration); err != nil {
		return err
	}

	return tx.Commit()
}

func rollbackMigration(db *sql.DB, migration string) error {
	// For simplicity, we'll just remove the migration record
	// In a real system, you'd want separate down migration files
	_, err := db.Exec("DELETE FROM schema_migrations WHERE version = $1", migration)
	return err
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

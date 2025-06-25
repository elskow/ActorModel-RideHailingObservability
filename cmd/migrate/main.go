package main

import (
	"actor-model-observability/internal/config"
	"actor-model-observability/internal/database"
	"actor-model-observability/internal/logging"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/rubenv/sql-migrate"
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

	// Initialize logger
	logger, err := logging.NewLogger(&cfg.Logging)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	// Connect to database using sqlx
	db, err := database.NewPostgresConnection(&cfg.Database, logger)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Create migration source from migrations directory
	migrations := &migrate.FileMigrationSource{
		Dir: "migrations",
	}

	// Execute command
	switch *command {
	case "up":
		if err := migrateUp(db, migrations, *steps); err != nil {
			log.Fatalf("Migration up failed: %v", err)
		}
	case "down":
		if err := migrateDown(db, migrations, *steps); err != nil {
			log.Fatalf("Migration down failed: %v", err)
		}
	case "status":
		if err := migrationStatus(db, migrations); err != nil {
			log.Fatalf("Migration status failed: %v", err)
		}
	default:
		log.Fatalf("Unknown command: %s", *command)
	}
}

func migrateUp(db *database.PostgresDB, migrations *migrate.FileMigrationSource, steps int) error {
	var n int
	var err error

	if steps > 0 {
		n, err = migrate.ExecMax(db.DB.DB, "postgres", migrations, migrate.Up, steps)
	} else {
		n, err = migrate.Exec(db.DB.DB, "postgres", migrations, migrate.Up)
	}

	if err != nil {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	if n == 0 {
		log.Println("No pending migrations")
	} else {
		log.Printf("Successfully applied %d migrations", n)
	}

	return nil
}

func migrateDown(db *database.PostgresDB, migrations *migrate.FileMigrationSource, steps int) error {
	if steps == 0 {
		steps = 1 // Default to rolling back 1 migration
	}

	n, err := migrate.ExecMax(db.DB.DB, "postgres", migrations, migrate.Down, steps)
	if err != nil {
		return fmt.Errorf("failed to rollback migrations: %w", err)
	}

	if n == 0 {
		log.Println("No migrations to rollback")
	} else {
		log.Printf("Successfully rolled back %d migrations", n)
	}

	return nil
}

func migrationStatus(db *database.PostgresDB, migrations *migrate.FileMigrationSource) error {
	// Get all migrations
	allMigrations, err := migrations.FindMigrations()
	if err != nil {
		return fmt.Errorf("failed to find migrations: %w", err)
	}

	// Get applied migrations
	records, err := migrate.GetMigrationRecords(db.DB.DB, "postgres")
	if err != nil {
		return fmt.Errorf("failed to get migration records: %w", err)
	}

	// Create a map of applied migrations for quick lookup
	applied := make(map[string]*migrate.MigrationRecord)
	for _, record := range records {
		applied[record.Id] = record
	}

	// Print status
	fmt.Printf("%-20s | %-10s | %s\n", "MIGRATION", "STATUS", "APPLIED AT")
	fmt.Println(strings.Repeat("-", 60))

	for _, migration := range allMigrations {
		if record, exists := applied[migration.Id]; exists {
			fmt.Printf("%-20s | %-10s | %s\n", migration.Id, "applied", record.AppliedAt.Format("2006-01-02 15:04:05"))
		} else {
			fmt.Printf("%-20s | %-10s | %s\n", migration.Id, "pending", "")
		}
	}

	return nil
}

// Helper function to check if migrations directory exists
func init() {
	if _, err := os.Stat("migrations"); os.IsNotExist(err) {
		log.Println("Warning: migrations directory not found. Make sure to run this command from the project root.")
	}
}

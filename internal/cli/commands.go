package cli

import (
	"crypto/sha256"
	"db-pivot/internal/config"
	"db-pivot/internal/db"
	"db-pivot/internal/diff"
	"db-pivot/internal/migration"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
)

var (
    dbmsFlag     string
    connFlag     string
    snapshotDir  string
    migrationDir string
)

var rootCmd = &cobra.Command{
    Use:   "dbpivot",
    Short: "Database schema migration tool",
}

func Execute() {
    if err := rootCmd.Execute(); err != nil {
        log.Fatalf("Error executing command: %v", err)
    }
}

func init() {
    rootCmd.AddCommand(initCmd)
    rootCmd.AddCommand(snapshotCmd)
    rootCmd.AddCommand(diffCmd)
    rootCmd.AddCommand(migrateCmd)
    rootCmd.AddCommand(applyCmd)
    rootCmd.AddCommand(rollbackCmd)

    initCmd.Flags().StringVarP(&dbmsFlag, "dbms", "d", "mysql", "Database management system (e.g., mysql)")
    initCmd.Flags().StringVarP(&connFlag, "connection", "c", "", "Database connection string (required)")
    initCmd.Flags().StringVarP(&snapshotDir, "snapshot-dir", "s", ".schema_manager/snapshots", "Directory for snapshots")
    initCmd.Flags().StringVarP(&migrationDir, "migration-dir", "m", ".schema_manager/migrations", "Directory for migrations")
    initCmd.MarkFlagRequired("connection")
}

var initCmd = &cobra.Command{
    Use:   "init",
    Short: "Initialize schema manager",
    Run: func(cmd *cobra.Command, args []string) {
        if err := createDirectories(); err != nil {
            log.Fatalf("Failed to create directories: %v", err)
        }
        cfg := config.Config{
            DBMS:         dbmsFlag,
            Connection:   connFlag,
            SnapshotDir:  snapshotDir,
            MigrationDir: migrationDir,
        }
        if err := config.InitConfig(cfg); err != nil {
            log.Fatalf("Failed to initialize config: %v", err)
        }
        dbManager, err := db.NewDBManager(dbmsFlag, connFlag)
        if err != nil {
            log.Fatalf("Failed to connect to database: %v", err)
        }
        if err := dbManager.InitVersionTable(); err != nil {
            log.Fatalf("Failed to initialize version table: %v", err)
        }
        log.Println("Schema manager initialized successfully")
    },
}

var snapshotCmd = &cobra.Command{
    Use:   "snapshot",
    Short: "Capture current database schema",
    Run: func(cmd *cobra.Command, args []string) {
        cfg, err := config.LoadConfig()
        if err != nil {
            log.Fatalf("Failed to load config: %v", err)
        }
        dbManager, err := db.NewDBManager(cfg.DBMS, cfg.Connection)
        if err != nil {
            log.Fatalf("Failed to connect to database: %v", err)
        }
        if err := dbManager.CaptureSnapshot(cfg.SnapshotDir); err != nil {
            log.Fatalf("Failed to capture snapshot: %v", err)
        }
        log.Println("Schema snapshot captured successfully")
    },
}

var diffCmd = &cobra.Command{
    Use:   "diff",
    Short: "Compare current schema with previous snapshot",
    Run: func(cmd *cobra.Command, args []string) {
        cfg, err := config.LoadConfig()
        if err != nil {
            log.Fatalf("Failed to load config: %v", err)
        }
        prevSnapshot, err := loadPreviousSnapshot(cfg.SnapshotDir)
        if err != nil {
            log.Fatalf("Failed to load previous snapshot: %v", err)
        }
        dbManager, err := db.NewDBManager(cfg.DBMS, cfg.Connection)
        if err != nil {
            log.Fatalf("Failed to connect to database: %v", err)
        }
        currSchema, err := dbManager.GetSchema()
        if err != nil {
            log.Fatalf("Failed to capture current schema: %v", err)
        }
        strategy := &diff.DefaultDiffStrategy{}
        changes, err := strategy.Compare(prevSnapshot, currSchema)
        if err != nil {
            log.Fatalf("Failed to compare schemas: %v", err)
        }
        if len(changes) == 0 {
            log.Println("No changes detected")
        } else {
            log.Println("Detected changes:")
            for _, change := range changes {
                log.Printf("- %s %s: %s", change.Type, change.Object, change.Detail)
            }
        }
    },
}

var migrateCmd = &cobra.Command{
    Use:   "migrate",
    Short: "Generate migration script based on schema changes",
    Run: func(cmd *cobra.Command, args []string) {
        cfg, err := config.LoadConfig()
        if err != nil {
            log.Fatalf("Failed to load config: %v", err)
        }
        prevSnapshot, err := loadPreviousSnapshot(cfg.SnapshotDir)
        if err != nil {
            log.Fatalf("Failed to load previous snapshot: %v", err)
        }
        dbManager, err := db.NewDBManager(cfg.DBMS, cfg.Connection)
        if err != nil {
            log.Fatalf("Failed to connect to database: %v", err)
        }
        currSchema, err := dbManager.GetSchema()
        if err != nil {
            log.Fatalf("Failed to capture current schema: %v", err)
        }
        strategy := &diff.DefaultDiffStrategy{}
        changes, err := strategy.Compare(prevSnapshot, currSchema)
        if err != nil {
            log.Fatalf("Failed to compare schemas: %v", err)
        }
        if len(changes) == 0 {
            log.Println("No migrations needed")
            return
        }
        mig, err := migration.GenerateMigration(changes, cfg.MigrationDir)
        if err != nil {
            log.Fatalf("Failed to generate migration: %v", err)
        }
        log.Printf("Migration script generated: %s_migration.sql", mig.Version)
    },
}

var applyCmd = &cobra.Command{
    Use:   "apply",
    Short: "Apply pending migrations",
    Run: func(cmd *cobra.Command, args []string) {
        cfg, err := config.LoadConfig()
        if err != nil {
            log.Fatalf("Failed to load config: %v", err)
        }

        dbManager, err := db.NewDBManager(cfg.DBMS, cfg.Connection)
        if err != nil {
            log.Fatalf("Failed to connect to database: %v", err)
        }

        if err := applyMigrations(dbManager, cfg.MigrationDir); err != nil {
            log.Fatalf("Failed to apply migrations: %v", err)
        }

        if err := dbManager.CaptureSnapshot(cfg.SnapshotDir); err != nil {
            log.Fatalf("Failed to capture post-migration snapshot: %v", err)
        }

        log.Println("All migrations applied successfully")
    },
}


var rollbackCmd = &cobra.Command{
    Use:   "rollback",
    Short: "Rollback the last applied migration",
    Run: func(cmd *cobra.Command, args []string) {
        cfg, err := config.LoadConfig()
        if err != nil {
            log.Fatalf("Failed to load config: %v", err)
        }

        dbManager, err := db.NewDBManager(cfg.DBMS, cfg.Connection)
        if err != nil {
            log.Fatalf("Failed to connect to database: %v", err)
        }

        lastVersion, err := dbManager.GetLastAppliedMigration()
        if err != nil {
            log.Fatalf("Failed to get last applied migration: %v", err)
        }
        if lastVersion == "" {
            log.Println("No migrations to rollback")
            return
        }

        scriptPath := filepath.Join(cfg.MigrationDir, fmt.Sprintf("%s_migration.sql", lastVersion))
        script, err := os.ReadFile(scriptPath)
        if err != nil {
            log.Fatalf("Failed to read migration file %s: %v", lastVersion, err)
        }

        mig := migration.Migration{
            Version:  lastVersion,
            UpScript: string(script),
            Checksum: fmt.Sprintf("%x", sha256.Sum256(script)),
        }
 
        if err := migration.RollbackMigration(dbManager, mig); err != nil {
            log.Fatalf("Failed to rollback migration %s: %v", lastVersion, err)
        }
 
        if err := dbManager.CaptureSnapshot(cfg.SnapshotDir); err != nil {
            log.Fatalf("Failed to capture post-rollback snapshot: %v", err)
        }

        log.Printf("Migration %s rolled back successfully", lastVersion)
    },
}

func createDirectories() error {
    dirs := []string{".schema_manager", snapshotDir, migrationDir}
    for _, dir := range dirs {
        if err := os.MkdirAll(dir, 0755); err != nil {
            return err
        }
    }
    return nil
}

func loadPreviousSnapshot(snapshotDir string) (map[string]interface{}, error) {
    files, err := os.ReadDir(snapshotDir)
    if err != nil {
        return nil, err
    }
    if len(files) == 0 {
        return make(map[string]interface{}), nil
    }
    sort.Slice(files, func(i, j int) bool {
        return files[i].Name() > files[j].Name()
    })
    latest := files[0].Name()
    data, err := os.ReadFile(filepath.Join(snapshotDir, latest))
    if err != nil {
        return nil, err
    }
    var schema map[string]interface{}
    if err := json.Unmarshal(data, &schema); err != nil {
        return nil, err
    }
    return schema, nil
}

func applyMigrations(dbManager *db.DBManager, migrationDir string) error {
    files, err := os.ReadDir(migrationDir)
    if err != nil {
        return fmt.Errorf("failed to read migration directory: %v", err)
    }

    var migrationFiles []string
    for _, file := range files {
        if strings.HasSuffix(file.Name(), ".sql") {
            migrationFiles = append(migrationFiles, file.Name())
        }
    }

    sort.Strings(migrationFiles)

    for _, file := range migrationFiles {
        version := strings.TrimSuffix(file, "_migration.sql")
        applied, err := dbManager.IsMigrationApplied(version)
        if err != nil {
            return err
        }
        if applied {
            log.Printf("Migration %s already applied, skipping", version)
            continue
        }

        script, err := os.ReadFile(filepath.Join(migrationDir, file))
        if err != nil {
            return fmt.Errorf("failed to read migration file %s: %v", file, err)
        }

        mig := migration.Migration{
            Version:  version,
            UpScript: string(script),
            Checksum: fmt.Sprintf("%x", sha256.Sum256(script)),
        }

        if err := migration.ApplyMigration(dbManager, mig); err != nil {
            return fmt.Errorf("failed to apply migration %s: %v", version, err)
        }

        log.Printf("Migration %s applied successfully", version)
    }

    return nil
}

 
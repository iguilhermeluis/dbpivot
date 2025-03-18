package db

import (
	"db-pivot/internal/adapters"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type DBManager struct {
    adapter adapters.DBAdapter
}

func NewDBManager(dbms, conn string) (*DBManager, error) {
    factory := &adapters.AdapterFactory{}
    adapter := factory.CreateAdapter(dbms, conn)
    if adapter == nil {
        return nil, fmt.Errorf("unsupported DBMS: %s", dbms)
    }
    if err := adapter.Connect(); err != nil {
        return nil, err
    }
    return &DBManager{adapter: adapter}, nil
}

func (d *DBManager) InitVersionTable() error {
    query := `
        CREATE TABLE IF NOT EXISTS schema_migrations (
            version VARCHAR(50) PRIMARY KEY,
            applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            description TEXT,
            checksum VARCHAR(64)
        )`
    return d.adapter.ApplyMigration(query)
}

func (d *DBManager) CaptureSnapshot(snapshotDir string) error {
    schema, err := d.adapter.GetSchema()
    if err != nil {
        return fmt.Errorf("failed to get schema: %v", err)
    }
    data, err := json.MarshalIndent(schema, "", "  ")
    if err != nil {
        return fmt.Errorf("failed to marshal schema: %v", err)
    }
    timestamp := time.Now().Format("20060102150405")
    filename := filepath.Join(snapshotDir, fmt.Sprintf("snapshot_%s.json", timestamp))
    if err := os.WriteFile(filename, data, 0644); err != nil {
        return fmt.Errorf("failed to write snapshot: %v", err)
    }
    return nil
}

func (d *DBManager) GetSchema() (map[string]interface{}, error) {
    return d.adapter.GetSchema()
}

func (d *DBManager) ApplyMigration(script string) error {
    return d.adapter.ApplyMigration(script)
}

func (d *DBManager) IsMigrationApplied(version string) (bool, error) {
    query := `SELECT COUNT(*) FROM schema_migrations WHERE version = ?`
    var count int
    err := d.adapter.QueryRow(query, version).Scan(&count)
    if err != nil {
        return false, fmt.Errorf("failed to check if migration %s is applied: %v", version, err)
    }
    return count > 0, nil
}

func (d *DBManager) GetLastAppliedMigration() (string, error) {
    query := `SELECT version FROM schema_migrations ORDER BY applied_at DESC LIMIT 1`
    var version string
    row := d.adapter.QueryRow(query)
    err := row.Scan(&version)
    if err != nil {
        if err.Error() == "sql: no rows in result set" {
            return "", nil  
        }
        return "", fmt.Errorf("failed to get last applied migration: %v", err)
    }
    return version, nil
}
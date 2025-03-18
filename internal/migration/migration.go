package migration

import (
	"crypto/sha256"
	"db-pivot/internal/db"
	"db-pivot/internal/diff"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Migration struct {
	Version    string
	UpScript   string
	DownScript string
	Checksum   string
}

func GenerateMigration(changes []diff.Change, migrationDir string) (Migration, error) {
	timestamp := time.Now().Format("20060102150405")
	version := timestamp
	filename := filepath.Join(migrationDir, fmt.Sprintf("%s_migration.sql", version))

	var upScript, downScript strings.Builder
	upScript.WriteString("-- Up migration\n")
	downScript.WriteString("-- Down migration\n")

	for _, change := range changes {
		switch change.Type {
		case "add":
			if strings.HasPrefix(change.Object, "table:") {
				table := strings.TrimPrefix(change.Object, "table:")
				tableDefinition := strings.TrimSpace(change.Detail)
			 	if tableDefinition == "" || strings.ToLower(tableDefinition) == "table added" {
			 		tableDefinition = "id INT AUTO_INCREMENT PRIMARY KEY"
				}
				upScript.WriteString(fmt.Sprintf("CREATE TABLE %s (\n%s\n);\n", table, tableDefinition))
				downScript.WriteString(fmt.Sprintf("DROP TABLE %s;\n", table))
			} else if strings.HasPrefix(change.Object, "column:") {
				parts := strings.Split(strings.TrimPrefix(change.Object, "column:"), ".")
				if len(parts) != 2 {
					return Migration{}, fmt.Errorf("formato inválido para coluna: %s", change.Object)
				}
				table, column := parts[0], parts[1]
				colType := extractType(change.Detail)
				upScript.WriteString(fmt.Sprintf("ALTER TABLE %s ADD %s %s;\n", table, column, colType))
				downScript.WriteString(fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s;\n", table, column))
			}
		case "remove":
			if strings.HasPrefix(change.Object, "table:") {
				table := strings.TrimPrefix(change.Object, "table:")
				upScript.WriteString(fmt.Sprintf("DROP TABLE %s;\n", table))
			 } else if strings.HasPrefix(change.Object, "column:") {
				parts := strings.Split(strings.TrimPrefix(change.Object, "column:"), ".")
				if len(parts) != 2 {
					return Migration{}, fmt.Errorf("formato inválido para coluna: %s", change.Object)
				}
				table, column := parts[0], parts[1]
				upScript.WriteString(fmt.Sprintf("ALTER TABLE %s DROP COLUMN %s;\n", table, column))
			 }
		case "modify":
			if strings.HasPrefix(change.Object, "column:") {
				parts := strings.Split(strings.TrimPrefix(change.Object, "column:"), ".")
				if len(parts) != 2 {
					return Migration{}, fmt.Errorf("formato inválido para coluna: %s", change.Object)
				}
				table, column := parts[0], parts[1]
				newType := extractType(change.Detail)
				upScript.WriteString(fmt.Sprintf("ALTER TABLE %s MODIFY %s %s;\n", table, column, newType))
				oldType := extractOldType(change.Detail)
				downScript.WriteString(fmt.Sprintf("ALTER TABLE %s MODIFY %s %s;\n", table, column, oldType))
			}
		default:
			return Migration{}, fmt.Errorf("tipo de mudança não suportado: %s", change.Type)
		}
	}

	content := upScript.String() + "\n" + downScript.String()
	checksum := fmt.Sprintf("%x", sha256.Sum256([]byte(content)))
	if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
		return Migration{}, fmt.Errorf("falha ao escrever o arquivo de migração: %v", err)
	}

	return Migration{
		Version:    version,
		UpScript:   upScript.String(),
		DownScript: downScript.String(),
		Checksum:   checksum,
	}, nil
}

func RollbackMigration(dbManager *db.DBManager, mig Migration) error {
    lines := strings.Split(mig.UpScript+"\n"+mig.DownScript, "\n")
    var downScript strings.Builder
    inDownSection := false

    for _, line := range lines {
        if strings.HasPrefix(line, "-- Down migration") {
            inDownSection = true
            continue
        }
        if inDownSection && strings.TrimSpace(line) != "" {
            downScript.WriteString(line + "\n")
        }
    }

    if err := dbManager.ApplyMigration(downScript.String()); err != nil {
        return fmt.Errorf("falha ao aplicar a migração down %s: %v", mig.Version, err)
    }

    delScript := fmt.Sprintf("DELETE FROM schema_migrations WHERE version = '%s'", mig.Version)
    if err := dbManager.ApplyMigration(delScript); err != nil {
        return fmt.Errorf("falha ao remover o registro da migração %s: %v", mig.Version, err)
    }
    return nil
}

func ApplyMigration(dbManager *db.DBManager, mig Migration) error {
	lines := strings.Split(mig.UpScript+"\n"+mig.DownScript, "\n")
	var upScript strings.Builder
	inUpSection := false

	for _, line := range lines {
		if strings.HasPrefix(line, "-- Up migration") {
			inUpSection = true
			continue
		}
		if strings.HasPrefix(line, "-- Down migration") {
			inUpSection = false
			break
		}
		if inUpSection {
			upScript.WriteString(line + "\n")
		}
	}

	if err := dbManager.ApplyMigration(upScript.String()); err != nil {
		return fmt.Errorf("falha ao aplicar a migração up %s: %v", mig.Version, err)
	}

	desc := fmt.Sprintf("Migration %s applied", mig.Version)
	regScript := fmt.Sprintf("INSERT INTO schema_migrations (version, description, checksum) VALUES ('%s', '%s', '%s')", mig.Version, desc, mig.Checksum)
	if err := dbManager.ApplyMigration(regScript); err != nil {
		return fmt.Errorf("falha ao registrar a migração %s: %v", mig.Version, err)
	}
	return nil
}



func extractType(detail string) string {
	parts := strings.Split(detail, " ")
	for i, part := range parts {
		if part == "type" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return "VARCHAR(255)"
}

func extractOldType(detail string) string {
	parts := strings.Split(detail, " ")
	for i, part := range parts {
		if part == "from" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return "VARCHAR(255)"
}

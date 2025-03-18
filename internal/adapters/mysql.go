package adapters

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

type MySQLAdapter struct {
    conn string
    db   *sql.DB
}

func NewMySQLAdapter(conn string) *MySQLAdapter {
    return &MySQLAdapter{conn: conn}
}

func (m *MySQLAdapter) Connect() error {
    db, err := sql.Open("mysql", m.conn)
    if err != nil {
        return err
    }
    m.db = db
    return db.Ping()
}

func (m *MySQLAdapter) GetSchema() (map[string]interface{}, error) {
    schema := make(map[string]interface{})

  
    tables, err := m.getTables()
    if err != nil {
        return nil, err
    }

 
    for _, table := range tables {
        columns, err := m.getColumns(table)
        if err != nil {
            return nil, err
        }
        schema[table] = map[string]interface{}{
            "columns": columns,
        }
    }

    return schema, nil
}

func (m *MySQLAdapter) ApplyMigration(script string) error {
    _, err := m.db.Exec(script)
    return err
}

func (m *MySQLAdapter) getTables() ([]string, error) {
    rows, err := m.db.Query("SHOW TABLES")
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var tables []string
    for rows.Next() {
        var table string
        if err := rows.Scan(&table); err != nil {
            return nil, err
        }
        tables = append(tables, table)
    }
    return tables, rows.Err()
}

func (m *MySQLAdapter) getColumns(table string) (map[string]interface{}, error) {
    rows, err := m.db.Query("SHOW COLUMNS FROM " + table)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    columns := make(map[string]interface{})
    for rows.Next() {
        var field, colType, null, key, defaultVal, extra sql.NullString
        if err := rows.Scan(&field, &colType, &null, &key, &defaultVal, &extra); err != nil {
            return nil, err
        }
        columns[field.String] = map[string]interface{}{
            "type":    colType.String,
            "null":    null.String == "YES",
            "key":     key.String,
            "default": defaultVal.String,
            "extra":   extra.String,
        }
    }
    return columns, rows.Err()
}

func (m *MySQLAdapter) QueryRow(query string, args ...interface{}) *sql.Row {
    return m.db.QueryRow(query, args...)
}
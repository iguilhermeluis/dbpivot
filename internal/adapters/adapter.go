package adapters

import "database/sql"

type DBAdapter interface {
    Connect() error
    GetSchema() (map[string]interface{}, error)
    ApplyMigration(script string) error
    QueryRow(query string, args ...interface{}) *sql.Row
}

type AdapterFactory struct{}

func (f *AdapterFactory) CreateAdapter(dbms string, conn string) DBAdapter {
    switch dbms {
    case "mysql":
        return NewMySQLAdapter(conn)
    default:
        return nil
    }
}
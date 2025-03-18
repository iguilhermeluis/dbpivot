# DB-Pivot: Database Migration Tool in Go

![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/iguilhermeluis/dbpivot)
![GitHub license](https://img.shields.io/github/license/iguilhermeluis/dbpivot)
![GitHub last commit](https://img.shields.io/github/last-commit/iguilhermeluis/dbpivot)
![Build Status](https://img.shields.io/github/actions/workflow/status/iguilhermeluis/dbpivot/go.yml?branch=main)

**DB-Pivot** is a lightweight CLI tool written in Go for database schema migrations. Capture schema snapshots, detect changes, generate migration scripts, and apply or revert them effortlessly. Built for MySQL with extensibility for other databases, it’s perfect for schema versioning and automation.

## Why DB-Pivot?

- **Simple Database Versioning**: Track schema changes with ease.
- **Automation for Teams**: Streamline MySQL migrations in CI/CD pipelines.
- **Safe Schema Management**: Revert changes securely with rollback support.

## Features

- **Snapshots**: Save database schema states as JSON.
- **Diff**: Compare schemas to detect table and column changes.
- **Migrate**: Generate SQL migration scripts (Up/Down).
- **Apply**: Execute schema migrations on MySQL.
- **Rollback**: Undo the last migration.
- **Extensible**: Add support for new DBMS.

## Installation

### Prerequisites

- [Go](https://golang.org/dl/) 1.21 or higher
- MySQL database (support for other DBMS planned)

### Steps

1. Clone the repository:
   ```bash
   git clone https://github.com/iguilhermeluis/dbpivot.git
   cd dbpivot
   ```

2. Build the project:
   ```bash
   go build ./cmd/dbpivot
   ```

3. (Optional) Install globally:
   ```bash
   go install ./cmd/dbpivot
   ```

The \`dbpivot\` binary is now ready to use!

## How to Use

### Initialize the Project

Set up the environment and create the migration control table:

```bash
./dbpivot init --dbms mysql --connection "user:password@tcp(localhost:3306)/dbname"
```

### Capture a Snapshot

Save the current schema state:

```bash
./dbpivot snapshot
```

### Detect Changes

Compare the current schema with the last snapshot:

```bash
./dbpivot diff
```

### Generate Migration

Create a migration script based on the changes:

```bash
./dbpivot migrate
```

### Apply Migrations

Run all pending migrations:

```bash
./dbpivot apply
```

### Revert a Migration

Undo the last applied migration:

```bash
./dbpivot rollback
```

### Practical Example

```bash
# Initialize
./dbpivot init --dbms mysql --connection "root:pass@tcp(127.0.0.1:3306)/mydb"

# Initial snapshot
./dbpivot snapshot

# Add a table
mysql -u root -p mydb -e "CREATE TABLE users (id INT PRIMARY KEY, name VARCHAR(100));"

# New snapshot and migration
./dbpivot snapshot
./dbpivot migrate

# Apply
./dbpivot apply

# Revert
./dbpivot rollback
```

## Project Structure


```
dbpivot/
├── cmd/          # CLI entry point
├── internal/     # Internal packages
│   ├── adapters/ # DBMS adapters
│   ├── cli/      # Command logic
│   ├── config/   # Configuration management
│   ├── db/       # Database interaction
│   ├── diff/     # Schema comparison
│   └── migration/# Migration generation and application
├── .gitignore
├── go.mod
├── go.sum
└── README.md
```
 
- **Migrations**: Stored in \`.schema_manager/migrations/\`.
- **Snapshots**: Stored in \`.schema_manager/snapshots/\`.

## Contributing

We love contributions! Here’s how to get started:

1. Fork the repository.
2. Create a branch for your change:
   ```bash
   git checkout -b my-contribution
   ```
3. Commit your changes:
   ```bash
   git commit -m "Add my contribution"
   ```
4. Push to your fork:
   ```bash
   git push origin my-contribution
   ```
5. Open a Pull Request.

### Guidelines

- Add tests for new features (in \`internal/<package>/test\`).
- Use \`gofmt\` to format code.
- Document public functions with clear comments.

## License

Distributed under the [MIT License](LICENSE). See the \`LICENSE\` file for details.

## Roadmap

- [ ] Support for PostgreSQL and SQLite.
- [ ] Apply/rollback multiple migrations in one command.
- [ ] Support for indexes and foreign keys.
- [ ] GitHub Actions integration for CI/CD.

## Support and Contact

Found a bug or have a suggestion? Open an [issue](https://github.com/iguilhermeluis/dbpivot/issues) or email [contato@iguilhermeluis.com](mailto:contato@iguilhermeluis.com).
`

## Tags
- Database Migration
- Schema Versioning
- MySQL
- Go CLI
- Open Source
package postgres

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"sort"
	"time"

	_ "github.com/lib/pq"
)

//go:embed migration/*.sql
var migrationFS embed.FS

type DB struct {
	db     *sql.DB
	ctx    context.Context
	cancel func()

	DSN string

	// Returns the current time. Defaults to time.Now().
	// Can be mocked for tests.
	Now func() time.Time
}

func NewDB(dsn string) *DB {
	db := &DB{
		DSN: dsn,
	}

	// TODO: can you use withcancelcause here?
	db.ctx, db.cancel = context.WithCancel(context.Background())
	return db
}

// Tx wraps the SQL Tx object to provide a timestamp at the start of the transaction.
type Tx struct {
	*sql.Tx
	db  *DB
	now time.Time
}

// NullTime represents a helper wrapper for time.Time. It automatically converts
// time fields to/from RFC 3339 format. Also supports NULL for zero time.
type NullTime time.Time

// Scan reads a time value from the database.
func (n *NullTime) Scan(value interface{}) error {
	if value == nil {
		*(*time.Time)(n) = time.Time{}
		return nil
	} else if value, ok := value.(string); ok {
		*(*time.Time)(n), _ = time.Parse(time.RFC3339, value)
		return nil
	}
	return fmt.Errorf("NullTime: cannot scan to time.Time: %T", value)
}

// Value formats a time value for the database.
func (n *NullTime) Value() (driver.Value, error) {
	if n == nil || (*time.Time)(n).IsZero() {
		return nil, nil
	}
	return (*time.Time)(n).UTC().Format(time.RFC3339), nil
}

// FormatLimitOffset returns a SQL string for a given limit & offset.
// Clauses are only added if limit and/or offset are greater than zero.
func FormatLimitOffset(limit, offset int) string {
	if limit > 0 && offset > 0 {
		return fmt.Sprintf(`LIMIT %d OFFSET %d`, limit, offset)
	} else if limit > 0 {
		return fmt.Sprintf(`LIMIT %d`, limit)
	} else if offset > 0 {
		return fmt.Sprintf(`OFFSET %d`, offset)
	}
	return ""
}

func (db *DB) Open() (err error) {
	if db.DSN == "" {
		return fmt.Errorf("dsn required")
	}

	if db.db, err = sql.Open("postgres", db.DSN); err != nil {
		return err
	}

	if err = db.migrate(); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}

	go db.monitor()

	return nil
}

// migrate sets up migration tracking and executes pending migration files.
//
// Migration files are embedded in the postgres/migration folder and are executed
// in lexigraphical order.
//
// Once a migration is run, its name is stored in the 'migrations' table so it
// is not re-executed. Migrations run in a transaction to prevent partial
// migrations.
func (db *DB) migrate() error {
	// Ensure the 'migrations' table exists so we don't duplicate migrations.
	if _, err := db.db.Exec(`CREATE TABLE IF NOT EXISTS migrations (name TEXT PRIMARY KEY);`); err != nil {
		return fmt.Errorf("cannot create migrations table: %w", err)
	}

	// Read migration files from our embedded file system.
	// This uses Go 1.16's 'embed' package.
	names, err := fs.Glob(migrationFS, "migration/*.sql")
	if err != nil {
		return err
	}
	sort.Strings(names)

	// Loop over all migration files and execute them in order.
	for _, name := range names {
		if err := db.migrateFile(name); err != nil {
			return fmt.Errorf("migration error: name=%q err=%w", name, err)
		}
	}
	return nil
}

// migrate runs a single migration file within a transaction. On success, the
// migration file name is saved to the "migrations" table to prevent re-running.
func (db *DB) migrateFile(name string) error {
	tx, err := db.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Ensure migration has not already been run.
	var n int
	if err := tx.QueryRow(`SELECT COUNT(*) FROM migrations WHERE name = $1`, name).Scan(&n); err != nil {
		return err
	} else if n != 0 {
		return nil // already run migration, skip
	}

	// Read and execute migration file.
	if buf, err := fs.ReadFile(migrationFS, name); err != nil {
		return err
	} else if _, err := tx.Exec(string(buf)); err != nil {
		return err
	}

	// Insert record into migrations to prevent re-running migration.
	if _, err := tx.Exec(`INSERT INTO migrations (name) VALUES ($1)`, name); err != nil {
		return err
	}

	return tx.Commit()
}

// Close closes the database connection.
func (db *DB) Close() error {
	// Cancel background context.
	db.cancel()

	// Close database.
	if db.db != nil {
		return db.db.Close()
	}
	return nil
}

// BeginTx starts a transaction and returns a wrapper Tx type. This type
// provides a reference to the database and a fixed timestamp at the start of
// the transaction. The timestamp allows us to mock time during tests as well.
func (db *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*Tx, error) {
	tx, err := db.db.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}

	// Return wrapper Tx that includes the transaction start time.
	return &Tx{
		Tx: tx,
		db: db,
	}, nil
}

// monitor runs in a goroutine and periodically calculates internal stats.
func (db *DB) monitor() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-db.ctx.Done():
			return
		case <-ticker.C:
		}

		if err := db.updateStats(db.ctx); err != nil {
			log.Printf("stats error: %s", err)
		}
	}
}

func (db *DB) updateStats(ctx context.Context) error {
	fmt.Println("updating stats")
	return nil
}

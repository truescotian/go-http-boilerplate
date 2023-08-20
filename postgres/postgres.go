package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

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

func (db *DB) Open() (err error) {
	if db.DSN == "" {
		return fmt.Errorf("dsn required")
	}

	if db.db, err = sql.Open("postgres", db.DSN); err != nil {
		return err
	}

	go db.monitor()

	return nil
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

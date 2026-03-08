package database

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// NewPostgresDB opens a connection pool to PostgreSQL.
func NewPostgresDB(connStr string) (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", connStr)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	return db, nil
}

// Migrate executes the provided schema SQL against the database.
func Migrate(db *sqlx.DB, schema string) error {
	_, err := db.Exec(schema)
	return err
}

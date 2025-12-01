package repositories

import (
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

type DatabaseRepository struct {
	db  *sql.DB
	dbx *sqlx.DB
}

func NewDatabaseRepository(dsn string) (*DatabaseRepository, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("не удалось открыть базу данных: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("не удалось подключиться к базе данных: %w", err)
	}

	dbx := sqlx.NewDb(db, "pgx")

	return &DatabaseRepository{db: db, dbx: dbx}, nil
}

func (dr *DatabaseRepository) Ping() error {
	return dr.db.Ping()
}

func (dr *DatabaseRepository) Close() error {
	return dr.db.Close()
}

func (dr *DatabaseRepository) GetDB() *sql.DB {
	return dr.db
}

func (dr *DatabaseRepository) GetDBX() *sqlx.DB {
	return dr.dbx
}

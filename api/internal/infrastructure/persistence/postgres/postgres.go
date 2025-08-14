/*
* This file will create the connection to a sql database
* migration files will be converted to binary and executed
* Will wrap the connection with sqlx library (more utility)
* return a connection to that sqlx db for interacting with the database
 */

package postgres

import (
	"database/sql"
	"embed"
	"fmt"
	"strings"

	"github.com/AliTSayyed/VULX-AI-Website-Builder/api/internal/config"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	migrate "github.com/rubenv/sql-migrate"
)

//go:embed migrations/*.sql
var dbMigrations embed.FS

func NewDb(cfg config.Db) *sqlx.DB {
	dsn := fmt.Sprintf(
		"host=%s port=5432 dbname=%s user=%s password=%s sslmode=disable",
		cfg.Host,
		cfg.Name,
		cfg.User,
		strings.TrimSpace(cfg.Password),
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		panic(fmt.Errorf("failed to connect to database: %w", err))
	}

	migrations := migrate.EmbedFileSystemMigrationSource{
		FileSystem: dbMigrations,
		Root:       "migrations",
	}

	_, err = migrate.Exec(db, "postgres", migrations, migrate.Up)
	if err != nil {
		panic(fmt.Errorf("failed to run migrations: %w", err))
	}

	dbx := sqlx.NewDb(db, "postgres")
	return dbx
}

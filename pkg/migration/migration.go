package migrations

import (
	"database/sql"
	"fmt"

	"github.com/pressly/goose"

	"github.com/JulianaOsi/medhelp/pkg/config"
)

func run(db *sql.DB) error {
	if err := goose.Up(db, "."); err != nil {
		return fmt.Errorf("goose: up failed: %v", err)
	}

	return nil
}

func UpMigrations(conf *config.Config) error {
	db, err := sql.Open("postgres", conf.DB.ToString())
	if err != nil {
		return fmt.Errorf("goose: failed to open DB: %v\n", err)
	}

	return run(db)
}

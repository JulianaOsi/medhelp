package migrations

import (
	"database/sql"
	"log"

	"github.com/pressly/goose"
)

func Run(db *sql.DB) {
	if err := goose.Up(db, "."); err != nil {
		log.Fatalf("goose up failed: %v", err)
	}
}

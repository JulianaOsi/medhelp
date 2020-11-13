package main

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"

	"github.com/JulianaOsi/medhelp/pkg/config"
	migrations "github.com/JulianaOsi/medhelp/pkg/migration"
)

func main() {
	conf := config.ReadConfig()

	db, err := sql.Open("postgres", conf.DB.ToString())
	if err != nil {
		log.Fatalf("goose: failed to open DB: %v\n", err)
	}

	migrations.Run(db)
}

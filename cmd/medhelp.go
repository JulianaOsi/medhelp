package main

import (
	"log"

	_ "github.com/lib/pq"

	"github.com/JulianaOsi/medhelp/pkg/config"
	migrations "github.com/JulianaOsi/medhelp/pkg/migration"
	"github.com/JulianaOsi/medhelp/pkg/server"
	"github.com/JulianaOsi/medhelp/pkg/store"
)

func main() {
	conf := config.ReadConfig()

	err := migrations.UpMigrations(conf)
	if err != nil {
		log.Fatalf("failed to update migrations: %v\n", err)
	}

	if err := store.InitDB(conf.DB); err != nil {
		log.Fatalf("failed to create store: %v\n", err)
	}

	server.LaunchServer()
}

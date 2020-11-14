package main

import (
	"context"
	"fmt"
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

	s, err := store.New(conf.DB)
	if err != nil {
		log.Fatalf("failed to create store: %v\n", err)
	}

	analysis, err := s.GetAnalysisByDirectionId(context.Background(), "0")
	if err != nil {
		log.Fatalf("get analysis failed: %v\n", err)
	}

	for _, a := range analysis {
		fmt.Println(a.Id, a.Name, a.IsChecked)
	}

	server.LaunchServer()
}

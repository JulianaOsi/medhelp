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

	directions, err := s.GetDirections(context.Background())
	if err != nil {
		log.Fatalf("get directions failed: %v\n", err)
	}

	for _, d := range directions {
		fmt.Println(
			d.Id, d.PatientFirstName, d.PatientLastName,
			d.PatientPolicyNumber, d.PatientTel, d.Date,
			d.IcdCode, d.MedicalOrganization, d.Status,
		)
	}

	server.LaunchServer()
}

package config

import (
	"github.com/JulianaOsi/medhelp/pkg/store"
)

var SigningKey = []byte("")
var DBPassword = ""

type Config struct {
	DB *store.ConfigDB
}

func ReadConfig() *Config {
	db := &store.ConfigDB{
		Host:     "127.0.0.1",
		Port:     "5432",
		Name:     "medhelp",
		User:     "postgres",
		Password: DBPassword,
	}

	return &Config{DB: db}
}

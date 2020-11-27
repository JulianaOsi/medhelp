package store

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4/pgxpool"
)

var DB *Store

type Store struct {
	connPool *pgxpool.Pool
}

type ConfigDB struct {
	Host     string
	Port     string
	Name     string
	User     string
	Password string
}

func InitDB(config *ConfigDB) error {
	pool, err := pgxpool.Connect(context.Background(), config.ToString())
	if err != nil {
		return err
	}

	DB = &Store{connPool: pool}
	return nil
}

func (c *ConfigDB) ToString() string {
	if c.Password != "" {
		return fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=disable search_path=public",
			c.Host, c.Port, c.Name, c.User, c.Password)
	}
	return fmt.Sprintf("host=%s port=%s dbname=%s user=%s sslmode=disable search_path=public",
		c.Host, c.Port, c.Name, c.User)
}

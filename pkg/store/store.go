package store

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4/pgxpool"
)

type Store struct {
	connPool *pgxpool.Pool
}

type ConfigDB struct {
	Host string
	Port string
	Name string
	User string
}

func New(config *ConfigDB) (*Store, error) {
	pool, err := pgxpool.Connect(context.Background(), config.ToString())
	if err != nil {
		return nil, err
	}
	return &Store{connPool: pool}, nil
}

func (c *ConfigDB) ToString() string {
	return fmt.Sprintf("host=%s port=%s dbname=%s user=%s sslmode=disable search_path=public",
		c.Host, c.Port, c.Name, c.User)
}

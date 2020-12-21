package store

import (
	"context"
	"fmt"

	"github.com/doug-martin/goqu/v9"
	"github.com/jackc/pgx/v4"
)

type Doctor struct {
	Id        int    `json:"id"`
	Name      string `json:"name"`
	Specialty string `json:"specialty"`
}

type NewDoctor struct {
	Name      string `json:"name"`
	Specialty string `json:"specialty"`
}

func (s *Store) GetDoctor(ctx context.Context, name string, specialty string) (*Doctor, error) {
	sql, _, err := goqu.Select("id", "name", "specialty").
		From("doctor").
		Where(goqu.C("name").Eq(name), goqu.C("specialty").Eq(specialty)).
		ToSQL()
	if err != nil {
		return nil, fmt.Errorf("sql query build failed: %v", err)
	}

	rows, err := s.connPool.Query(ctx, sql)
	if err != nil {
		return nil, fmt.Errorf("execute a query failed: %v", err)
	}
	defer rows.Close()

	var doctors []*Doctor

	for rows.Next() {
		doctor, err := readDoctor(rows)
		if err != nil {
			return nil, fmt.Errorf("read direction failed: %v", err)
		}
		doctors = append(doctors, doctor)
	}

	if len(doctors) == 0 {
		return nil, nil
	}
	return doctors[0], nil
}

func (s *Store) AddDoctor(ctx context.Context, doctor NewDoctor) (*int, error) {
	sql, _, err := goqu.Insert("doctor").
		Rows(goqu.Record{
			"name":      doctor.Name,
			"specialty": doctor.Specialty,
		}).
		OnConflict(goqu.DoNothing()).
		ToSQL()
	if err != nil {
		return nil, fmt.Errorf("sql query build failed: %v", err)
	}
	if _, err := s.connPool.Exec(ctx, sql); err != nil {
		return nil, fmt.Errorf("execute a query failed: %v", err)
	}

	newDoctor, err := s.GetDoctor(context.Background(), doctor.Name, doctor.Specialty)
	if err != nil {
		return nil, fmt.Errorf("failed to get doctor: %v\n", err)
	}

	return &newDoctor.Id, nil
}

func readDoctor(row pgx.Row) (*Doctor, error) {
	var d Doctor

	err := row.Scan(
		&d.Id, &d.Name, &d.Specialty,
	)
	if err != nil {
		return nil, err
	}

	return &d, nil
}

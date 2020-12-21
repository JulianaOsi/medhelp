package store

import (
	"context"
	"fmt"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/jackc/pgx/v4"
)

type Patient struct {
	Id           int       `json:"id"`
	FirstName    string    `json:"firstName"`
	LastName     string    `json:"lastName"`
	BirthDate    time.Time `json:"birthDate"`
	PolicyNumber string    `json:"policyNumber"`
	Tel          string    `json:"tel"`
}

type NewPatient struct {
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	BirthDate    time.Time `json:"birth_date"`
	PolicyNumber string    `json:"policy_number"`
	Tel          string    `json:"tel"`
}

func (s *Store) GetPatient(ctx context.Context, lastName string, policyNumber string) (*Patient, error) {
	sql, _, err := goqu.Select("id", "first_name", "last_name", "birth_date", "policy_number", "tel").
		From("patient").
		Where(goqu.C("last_name").Eq(lastName), goqu.C("policy_number").Eq(policyNumber)).
		ToSQL()
	if err != nil {
		return nil, fmt.Errorf("sql query build failed: %v", err)
	}

	rows, err := s.connPool.Query(ctx, sql)
	if err != nil {
		return nil, fmt.Errorf("execute a query failed: %v", err)
	}
	defer rows.Close()

	var patients []*Patient

	for rows.Next() {
		patient, err := readPatient(rows)
		if err != nil {
			return nil, fmt.Errorf("read direction failed: %v", err)
		}
		patients = append(patients, patient)
	}

	if len(patients) == 0 {
		return nil, nil
	}
	return patients[0], nil
}

func (s *Store) AddPatient(ctx context.Context, patient NewPatient) (*int, error) {
	sql, _, err := goqu.Insert("patient").
		Rows(goqu.Record{
			"first_name":    patient.FirstName,
			"last_name":     patient.LastName,
			"birth_date":    patient.BirthDate,
			"policy_number": patient.PolicyNumber,
			"tel":           patient.Tel,
		}).
		OnConflict(goqu.DoNothing()).
		ToSQL()
	if err != nil {
		return nil, fmt.Errorf("sql query build failed: %v", err)
	}
	if _, err := s.connPool.Exec(ctx, sql); err != nil {
		return nil, fmt.Errorf("execute a query failed: %v", err)
	}

	newPatient, err := s.GetPatient(context.Background(), patient.LastName, patient.PolicyNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get patient: %v\n", err)
	}

	return &newPatient.Id, nil
}

func readPatient(row pgx.Row) (*Patient, error) {
	var p Patient

	err := row.Scan(
		&p.Id, &p.FirstName, &p.LastName,
		&p.BirthDate, &p.PolicyNumber, &p.Tel,
	)
	if err != nil {
		return nil, err
	}

	return &p, nil
}

package store

import (
	"context"
	"fmt"

	"github.com/doug-martin/goqu/v9"
	"github.com/jackc/pgx/v4"
)

type Patient struct {
	Id           int    `json:"id"`
	FirstName    string `json:"firstName"`
	LastName     string `json:"lastName"`
	PolicyNumber string `json:"policyNumber"`
	Tel          string `json:"tel"`
}

func (s *Store) GetPatient(ctx context.Context, lastName string, policyNumber string) (*Patient, error) {
	sql, _, err := goqu.Select("id", "first_name", "last_name", "policy_number", "tel").
		From("patient").
		Where(goqu.C("last_name").Eq(lastName), goqu.C("policy_number").Eq(policyNumber)).
		ToSQL()
	if err != nil {
		return nil, fmt.Errorf("GetPatientByLastNameAndPolicyNumber(): sql query build failed: %v", err)
	}

	rows, err := s.connPool.Query(ctx, sql)
	if err != nil {
		return nil, fmt.Errorf("GetPatientByLastNameAndPolicyNumber(): execute a query failed: %v", err)
	}
	defer rows.Close()

	var patients []*Patient

	for rows.Next() {
		patient, err := readPatient(rows)
		if err != nil {
			return nil, fmt.Errorf("GetPatientByLastNameAndPolicyNumber(): read direction failed: %v", err)
		}
		patients = append(patients, patient)
	}

	if len(patients) == 0 {
		return nil, nil
	}
	return patients[0], nil
}

func readPatient(row pgx.Row) (*Patient, error) {
	var p Patient

	err := row.Scan(
		&p.Id, &p.FirstName, &p.LastName,
		&p.PolicyNumber, &p.Tel,
	)
	if err != nil {
		return nil, err
	}

	return &p, nil
}

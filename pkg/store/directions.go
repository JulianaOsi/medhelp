package store

import (
	"context"
	"fmt"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/jackc/pgx/v4"
)

type Direction struct {
	Id                  int       `json:"id"`
	PatientFirstName    string    `json:"patientFirstName"`
	PatientLastName     string    `json:"patientLastName"`
	PatientPolicyNumber string    `json:"patientPolicyNumber"`
	PatientTel          string    `json:"patientTel"`
	Date                time.Time `json:"date"`
	IcdCode             string    `json:"icdCode"`
	MedicalOrganization string    `json:"medicalOrganization"`
	Status              int       `json:"status"`
}

func (s *Store) GetDirections(ctx context.Context) ([]*Direction, error) {
	sql, _, err := goqu.Select(
		"direction.id", "first_name", "last_name", "policy_number",
		"tel", "date", "icd_code", "medical_organization", "status",
	).
		From("direction").
		LeftJoin(
			goqu.T("patient"),
			goqu.On(goqu.Ex{
				"patient_id": goqu.I("patient.id"),
			}),
		).
		Order(goqu.C("date").Asc()).
		ToSQL()
	if err != nil {
		return nil, fmt.Errorf("sql query build failed: %v", err)
	}

	rows, err := s.connPool.Query(ctx, sql)
	if err != nil {
		return nil, fmt.Errorf("execute a query failed: %v", err)
	}
	defer rows.Close()

	var directions []*Direction

	for rows.Next() {
		direction, err := readDirection(rows)
		if err != nil {
			return nil, fmt.Errorf("read direction failed: %v", err)
		}
		directions = append(directions, direction)
	}

	return directions, nil
}

func (s *Store) GetDirectionsByPatientId(ctx context.Context, patientId string) ([]*Direction, error) {
	sql, _, err := goqu.Select(
		"direction.id", "first_name", "last_name", "policy_number",
		"tel", "date", "icd_code", "medical_organization", "status",
	).
		From("direction").
		LeftJoin(
			goqu.T("patient"),
			goqu.On(goqu.Ex{
				"patient_id": goqu.I("patient.id"),
			}),
		).
		Where(goqu.C("patient_id").Eq(patientId)).
		Order(goqu.C("date").Asc()).
		ToSQL()
	if err != nil {
		return nil, fmt.Errorf("sql query build failed: %v", err)
	}

	rows, err := s.connPool.Query(ctx, sql)
	if err != nil {
		return nil, fmt.Errorf("execute a query failed: %v", err)
	}
	defer rows.Close()

	var directions []*Direction

	for rows.Next() {
		direction, err := readDirection(rows)
		if err != nil {
			return nil, fmt.Errorf("read direction failed: %v", err)
		}
		directions = append(directions, direction)
	}

	return directions, nil
}

func (s *Store) SetDirectionStatus(ctx context.Context, directionId int, statusId int) error {
	sql, _, err := goqu.Update("direction").
		Set(goqu.Record{"status": statusId}).
		Where(goqu.C("id").Eq(directionId)).
		ToSQL()
	if err != nil {
		return fmt.Errorf("sql query build failed: %v", err)
	}

	_, err = s.connPool.Exec(ctx, sql)

	if err != nil {
		return fmt.Errorf("execute a query failed: %v", err)
	}

	return nil
}

func readDirection(row pgx.Row) (*Direction, error) {
	var d Direction

	err := row.Scan(
		&d.Id, &d.PatientFirstName, &d.PatientLastName,
		&d.PatientPolicyNumber, &d.PatientTel, &d.Date,
		&d.IcdCode, &d.MedicalOrganization, &d.Status,
	)
	if err != nil {
		return nil, err
	}

	return &d, nil
}

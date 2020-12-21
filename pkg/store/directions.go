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
	PatientBirthDate    time.Time `json:"patientBirthDate"`
	PatientPolicyNumber string    `json:"patientPolicyNumber"`
	PatientTel          string    `json:"patientTel"`
	DoctorName          string    `json:"doctorName"`
	DoctorSpecialty     string    `json:"doctorSpecialty"`
	Date                time.Time `json:"date"`
	IcdCode             string    `json:"icdCode"`
	MedicalOrganization string    `json:"medicalOrganization"`
	OrganizationContact string    `json:"organizationContact"`
	Justification       string    `json:"justification"`
	Status              int       `json:"status"`
}

type NewDirection struct {
	PatientId           int       `json:"patientId"`
	DoctorId            int       `json:"doctorId"`
	Date                time.Time `json:"date"`
	IcdCode             string    `json:"icdCode"`
	MedicalOrganization string    `json:"medicalOrganization"`
	OrganizationContact string    `json:"organizationContact"`
	Justification       string    `json:"justification"`
}

func (s *Store) AddDirection(ctx context.Context, direction NewDirection) error {
	sql, _, err := goqu.Insert("direction").
		Rows(goqu.Record{
			"patient_id":           direction.PatientId,
			"doctor_id":            direction.DoctorId,
			"date":                 direction.Date,
			"icd_code":             direction.IcdCode,
			"medical_organization": direction.MedicalOrganization,
			"organization_contact": direction.OrganizationContact,
			"justification":        direction.Justification,
		}).
		OnConflict(goqu.DoNothing()).
		ToSQL()
	if err != nil {
		return fmt.Errorf("sql query build failed: %v", err)
	}

	if _, err := s.connPool.Exec(ctx, sql); err != nil {
		return fmt.Errorf("execute a query failed: %v", err)
	}
	return nil
}

func (s *Store) GetDirections(ctx context.Context) ([]*Direction, error) {
	sql, _, err := goqu.Select(
		"direction.id", "first_name", "last_name", "birth_date", "policy_number", "tel", "name",
		"specialty", "date", "icd_code", "medical_organization", "organization_contact", "justification", "status",
	).
		From("direction").
		LeftJoin(
			goqu.T("patient"),
			goqu.On(goqu.Ex{
				"patient_id": goqu.I("patient.id"),
			}),
		).
		LeftJoin(
			goqu.T("doctor"),
			goqu.On(goqu.Ex{
				"doctor_id": goqu.I("doctor.id"),
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
		"direction.id", "first_name", "last_name", "birth_date", "policy_number", "tel", "name",
		"specialty", "date", "icd_code", "medical_organization", "organization_contact", "justification", "status",
	).
		From("direction").
		LeftJoin(
			goqu.T("patient"),
			goqu.On(goqu.Ex{
				"patient_id": goqu.I("patient.id"),
			}),
		).
		LeftJoin(
			goqu.T("doctor"),
			goqu.On(goqu.Ex{
				"doctor_id": goqu.I("doctor.id"),
			}),
		).
		Where(goqu.C("patient_id").Eq(patientId)).
		Order(goqu.C("date").Asc()).
		ToSQL()
	if err != nil {
		return nil, fmt.Errorf("GetDirectionsByPatientId(): sql query build failed: %v", err)
	}

	rows, err := s.connPool.Query(ctx, sql)
	if err != nil {
		return nil, fmt.Errorf("GetDirectionsByPatientId(): execute a query failed: %v", err)
	}
	defer rows.Close()

	var directions []*Direction

	for rows.Next() {
		direction, err := readDirection(rows)
		if err != nil {
			return nil, fmt.Errorf("GetDirectionsByPatientId(): read direction failed: %v", err)
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
		&d.Id, &d.PatientFirstName, &d.PatientLastName, &d.PatientBirthDate,
		&d.PatientPolicyNumber, &d.PatientTel, &d.DoctorName,
		&d.DoctorSpecialty, &d.Date, &d.IcdCode, &d.MedicalOrganization,
		&d.OrganizationContact, &d.Justification, &d.Status,
	)
	if err != nil {
		return nil, err
	}

	return &d, nil
}

package migrations

import (
	"database/sql"

	"github.com/pressly/goose"
)

func init() {
	goose.AddMigration(upAddDoctorIdToDirection, downAddDoctorIdToDirection)
}

func upAddDoctorIdToDirection(tx *sql.Tx) error {
	_, err := tx.Exec(`
ALTER TABLE direction ADD COLUMN doctor_id INT NOT NULL;
ALTER TABLE direction ADD FOREIGN KEY (doctor_id) REFERENCES doctor (id)
`)
	return err
}

func downAddDoctorIdToDirection(tx *sql.Tx) error {
	_, err := tx.Exec(`
ALTER TABLE direction DROP COLUMN doctor_id;
`)
	return err
}

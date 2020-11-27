package migrations

import (
	"database/sql"

	"github.com/pressly/goose"
)

func init() {
	goose.AddMigration(upAddDoctorTable, downAddDoctorTable)
}

func upAddDoctorTable(tx *sql.Tx) error {
	_, err := tx.Exec(`
CREATE TABLE IF NOT EXISTS doctor
(
    id 		  INT   PRIMARY KEY,
    name      TEXT  NOT NULL,
    specialty TEXT  NOT NULL
);
`)
	return err
}

func downAddDoctorTable(tx *sql.Tx) error {
	_, err := tx.Exec(`
DROP TABLE doctor CASCADE;
`)
	return err
}

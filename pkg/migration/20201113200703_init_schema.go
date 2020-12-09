package migrations

import (
	"database/sql"

	"github.com/pressly/goose"
)

func init() {
	goose.AddMigration(upInitSchema, downInitSchema)
}

func upInitSchema(tx *sql.Tx) error {
	_, err := tx.Exec(`
CREATE TABLE IF NOT EXISTS patient
(
    id            INT PRIMARY KEY,
    first_name    TEXT        NOT NULL,
    last_name     TEXT        NOT NULL,
    policy_number TEXT UNIQUE NOT NULL,
    tel           TEXT
);

CREATE TABLE IF NOT EXISTS direction
(
    id                   INT PRIMARY KEY,
    patient_id           INT      NOT NULL,
    date                 TIMESTAMP NOT NULL,
    icd_code             TEXT,
    medical_organization TEXT,
    status               INT DEFAULT 0,
    FOREIGN KEY (patient_id) REFERENCES patient (id)
);

CREATE TABLE IF NOT EXISTS analysis
(
    id       INT PRIMARY KEY,
    name     TEXT NOT NULL,
    icd_code TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS direction_analysis
(
    id           INT PRIMARY KEY,
    analysis_id  INT NOT NULL,
    direction_id INT NOT NULL,
    is_checked   BOOLEAN DEFAULT FALSE,
    FOREIGN KEY (analysis_id) REFERENCES analysis (id),
    FOREIGN KEY (direction_id) REFERENCES direction (id)
);

CREATE TABLE IF NOT EXISTS users
(
    id       	SERIAL PRIMARY KEY,
    username	TEXT NOT NULL,
    password	TEXT NOT NULL,
    salt		TEXT NOT NULL,
    role		TEXT NOT NULL
);
`)
	return err
}

func downInitSchema(tx *sql.Tx) error {
	_, err := tx.Exec(`
DROP TABLE direction CASCADE;
DROP TABLE analysis CASCADE;
DROP TABLE patient CASCADE;
DROP TABLE direction_analysis CASCADE;
DROP TABLE users CASCADE;

`)
	return err
}

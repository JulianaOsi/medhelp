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
    id 			  INT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    first_name    TEXT        NOT NULL,
    last_name     TEXT        NOT NULL,
    birth_date	  DATE		  NOT NULL,
    policy_number TEXT UNIQUE NOT NULL,
    tel           TEXT		  NOT NULL
);

CREATE TABLE IF NOT EXISTS doctor
(
    id 		  INT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    name      TEXT  NOT NULL,
    specialty TEXT  NOT NULL
);

CREATE TABLE IF NOT EXISTS direction
(
    id                   INT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    patient_id           INT NOT NULL,
    doctor_id			 INT NOT NULL,
    date                 TIMESTAMP NOT NULL,
    icd_code             TEXT,
    medical_organization TEXT,
    organization_contact TEXT,
    justification		 TEXT,
    status               INT DEFAULT 0,
    FOREIGN KEY (patient_id) REFERENCES patient (id),
    FOREIGN KEY (doctor_id) REFERENCES doctor (id)
);

CREATE TABLE IF NOT EXISTS icd_analysis
(
    id 			INT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    icd_code 	TEXT,
    analysis_id INT
);

CREATE TABLE IF NOT EXISTS analysis
(
    id       INT PRIMARY KEY,
    name     TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS files
(
    id           INT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    name  		 TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS direction_analysis
(
    id           INT PRIMARY KEY,
    analysis_id  INT NOT NULL,
    direction_id INT NOT NULL,
    is_checked   BOOLEAN DEFAULT FALSE,
    file_id		 INT,
    FOREIGN KEY (analysis_id) REFERENCES analysis (id),
    FOREIGN KEY (direction_id) REFERENCES direction (id),
    FOREIGN KEY (file_id) REFERENCES files (id)
);

CREATE TABLE IF NOT EXISTS users
(
    id       	INT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    username	TEXT NOT NULL,
    password	TEXT NOT NULL,
    salt		TEXT NOT NULL,
    role		TEXT NOT NULL,
    id_related	INT
);
`)
	return err
}

func downInitSchema(tx *sql.Tx) error {
	_, err := tx.Exec(`
DROP TABLE direction CASCADE;
DROP TABLE analysis CASCADE;
DROP TABLE patient CASCADE;
DROP TABLE doctor CASCADE;
DROP TABLE files CASCADE;
DROP TABLE direction_analysis CASCADE;
DROP TABLE users CASCADE;
`)
	return err
}

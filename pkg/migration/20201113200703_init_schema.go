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
    id            UUID PRIMARY KEY,
    first_name    TEXT        NOT NULL,
    last_name     TEXT        NOT NULL,
    policy_number TEXT UNIQUE NOT NULL,
    tel           TEXT
);

CREATE TABLE IF NOT EXISTS direction
(
    id                   UUID PRIMARY KEY,
    patient_id           UUID      NOT NULL,
    date                 TIMESTAMP NOT NULL,
    icd_code             TEXT,
    medical_organization TEXT,
    status               INT DEFAULT 0,
    FOREIGN KEY (patient_id) REFERENCES patient (id)
);

CREATE TABLE IF NOT EXISTS analysis
(
    id       UUID PRIMARY KEY,
    name     TEXT NOT NULL,
    icd_code TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS direction_analysis
(
    id           UUID PRIMARY KEY,
    analysis_id  UUID NOT NULL,
    direction_id UUID NOT NULL,
    is_checked   BOOLEAN DEFAULT FALSE,
    FOREIGN KEY (analysis_id) REFERENCES analysis (id),
    FOREIGN KEY (direction_id) REFERENCES direction (id)
);
`)
	return err
}

func downInitSchema(tx *sql.Tx) error {
	return nil
}

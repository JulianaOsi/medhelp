package migrations

import (
	"database/sql"

	"github.com/pressly/goose"
)

func init() {
	goose.AddMigration(upAddAnalysisFileTable, downAddAnalysisFileTable)
}

func upAddAnalysisFileTable(tx *sql.Tx) error {
	_, err := tx.Exec(`
CREATE TABLE IF NOT EXISTS analysis_file
(
    id           INT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    name  		 TEXT NOT NULL UNIQUE,
    direction_id INT NOT NULL,
    FOREIGN KEY (direction_id) REFERENCES direction (id)
);
`)
	return err
}

func downAddAnalysisFileTable(tx *sql.Tx) error {
	_, err := tx.Exec(`
DROP TABLE analysis_file CASCADE;
`)
	return err
}

package store

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"

	"github.com/doug-martin/goqu/v9"
)

var directory string

func init() {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	directory = wd + "/files/analysis/"
}

func (s *Store) UploadAnalysisFile(ctx context.Context, name string, file multipart.File, directionId string) error {
	filePath := directory + name
	if err := saveFile(filePath, file); err != nil {
		return err
	}

	sql, _, err := goqu.Insert("analysis_file").
		Rows(goqu.Record{"name": name, "direction_id": directionId}).
		ToSQL()
	if err != nil {
		os.Remove(filePath)
		return fmt.Errorf("sql query build failed: %v", err)
	}

	_, err = s.connPool.Exec(ctx, sql)
	if err != nil {
		os.Remove(filePath)
		return fmt.Errorf("execute a query failed: %v", err)
	}

	return nil
}

func saveFile(filePath string, file multipart.File) error {
	f, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer f.Close()

	_, err = io.Copy(f, file)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

package store

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"

	"github.com/doug-martin/goqu/v9"
	"github.com/jackc/pgx/v4"
)

var directory string

func init() {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	directory = wd + "/files/analysis/"
}

type AnalysisFile struct {
	Id          int            `json:"id"`
	FileName    string         `json:"fileName"`
	DirectionId int            `json:"directionId"`
	File        multipart.File `json:"file"`
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

func (s *Store) GetAnalysisFiles(ctx context.Context, directionId string) ([]*AnalysisFile, error) {

	sql, _, err := goqu.From("analysis_file").
		Where(goqu.C("direction_id").Eq(directionId)).
		ToSQL()
	if err != nil {
		return nil, fmt.Errorf("sql query build failed: %v", err)
	}

	rows, err := s.connPool.Query(ctx, sql)
	if err != nil {
		return nil, fmt.Errorf("execute a query failed: %v", err)
	}
	defer rows.Close()

	var analysisFiles []*AnalysisFile
	for rows.Next() {
		analysisFile, err := readAnalysisFile(rows)
		if err != nil {
			return nil, fmt.Errorf("read analysis file failed: %v", err)
		}
		f, err := os.Open(directory + analysisFile.FileName)
		if err != nil {
			return nil, fmt.Errorf("failed to open file: %w", err)
		}

		analysisFile.File = multipart.File(f)
		analysisFiles = append(analysisFiles, analysisFile)
	}

	return analysisFiles, nil
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

func readAnalysisFile(row pgx.Row) (*AnalysisFile, error) {
	var f AnalysisFile

	err := row.Scan(&f.Id, &f.FileName, &f.DirectionId)
	if err != nil {
		return nil, err
	}

	return &f, nil
}

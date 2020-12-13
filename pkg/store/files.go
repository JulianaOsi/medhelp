package store

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/jackc/pgx/v4"
)

var directory string

func init() {
	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	directory = wd + "/files/"
}

type File struct {
	Id   int    `json:"id"`
	Name string `json:"filename"`
}

func (s *Store) SaveFile(ctx context.Context, file multipart.File, filename string) (*int, error) {
	name := strconv.FormatInt(time.Now().Unix(), 10) + filepath.Ext(filename)
	filePath := directory + name
	if err := saveFile(filePath, file); err != nil {
		return nil, err
	}

	sql, _, err := goqu.Insert("files").
		Rows(goqu.Record{"name": name}).
		ToSQL()
	if err != nil {
		err = os.Remove(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to remove file: %v", err)
		}
		return nil, fmt.Errorf("sql query build failed: %v", err)
	}

	_, err = s.connPool.Exec(ctx, sql)
	if err != nil {
		err = os.Remove(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to remove file: %v", err)
		}
		return nil, fmt.Errorf("execute a query failed: %v", err)
	}

	sql, _, err = goqu.Select("id", "name").
		From("files").
		Where(goqu.C("name").Eq(name)).
		ToSQL()
	if err != nil {
		return nil, fmt.Errorf("sql query build failed: %v", err)
	}

	rows, err := s.connPool.Query(ctx, sql)
	if err != nil {
		return nil, fmt.Errorf("execute a query failed: %v", err)
	}
	defer rows.Close()

	var files []*File

	for rows.Next() {
		file, err := readFile(rows)
		if err != nil {
			return nil, fmt.Errorf("converting failed: %v", err)
		}
		files = append(files, file)
	}

	if len(files) != 0 {
		return &files[0].Id, nil
	}

	return nil, fmt.Errorf("failed to get file id: %v", err)
}

func (s *Store) GetFilepath(ctx context.Context, fileId int) (*string, error) {
	sql, _, err := goqu.Select("id", "name").
		From("files").
		Where(goqu.C("id").Eq(fileId)).
		ToSQL()
	if err != nil {
		return nil, fmt.Errorf("sql query build failed: %v", err)
	}

	rows, err := s.connPool.Query(ctx, sql)
	if err != nil {
		return nil, fmt.Errorf("execute a query failed: %v", err)
	}
	defer rows.Close()

	var files []*File

	for rows.Next() {
		file, err := readFile(rows)
		if err != nil {
			return nil, fmt.Errorf("converting failed: %v", err)
		}
		files = append(files, file)
	}

	if len(files) != 0 {
		var filePath = directory + files[0].Name
		return &filePath, nil
	}

	return nil, nil
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

func readFile(row pgx.Row) (*File, error) {
	var f File

	err := row.Scan(&f.Id, &f.Name)
	if err != nil {
		return nil, err
	}

	return &f, nil
}

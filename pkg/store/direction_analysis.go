package store

import (
	"context"
	"fmt"

	"github.com/doug-martin/goqu/v9"
	"github.com/jackc/pgx/v4"
)

type Analysis struct {
	Id          int    `json:"id"`
	Name        string `json:"name"`
	IsChecked   bool   `json:"isChecked"`
	FileId      *int   `json:"file_id"`
	DirectionId int    `json:"direction_id"`
}

func (s *Store) GetAnalysisByDirectionId(ctx context.Context, directionId int) ([]*Analysis, error) {
	sql, _, err := goqu.Select("direction_analysis.id", "name", "is_checked", "file_id", "direction_id").
		From("direction_analysis").
		Where(goqu.C("direction_id").Eq(directionId)).
		LeftJoin(
			goqu.T("analysis"),
			goqu.On(goqu.Ex{
				"analysis.id": goqu.I("direction_analysis.analysis_id"),
			}),
		).
		ToSQL()
	if err != nil {
		return nil, fmt.Errorf("GetAnalysisByDirectionId(): sql query build failed: %v", err)
	}

	rows, err := s.connPool.Query(ctx, sql)
	if err != nil {
		return nil, fmt.Errorf("GetAnalysisByDirectionId(): execute a query failed: %v", err)
	}
	defer rows.Close()

	var analysis []*Analysis

	for rows.Next() {
		newAnalysis, err := readAnalysis(rows)
		if err != nil {
			return nil, fmt.Errorf("GetAnalysisByDirectionId(): read analysis failed: %v", err)
		}
		analysis = append(analysis, newAnalysis)
	}

	return analysis, nil
}

func (s *Store) GetAnalysisById(ctx context.Context, id int) (*Analysis, error) {
	sql, _, err := goqu.Select("direction_analysis.id", "name", "is_checked", "file_id", "direction_id").
		From("direction_analysis").
		Where(goqu.L("\"direction_analysis\".\"id\"").Eq(id)).
		LeftJoin(
			goqu.T("analysis"),
			goqu.On(goqu.Ex{
				"analysis.id": goqu.I("direction_analysis.analysis_id"),
			}),
		).
		ToSQL()
	if err != nil {
		return nil, fmt.Errorf("sql query build failed: %v", err)
	}

	rows, err := s.connPool.Query(ctx, sql)
	if err != nil {
		return nil, fmt.Errorf("execute a query failed: %v", err)
	}
	defer rows.Close()

	var analysis []*Analysis

	for rows.Next() {
		newAnalysis, err := readAnalysis(rows)
		if err != nil {
			return nil, fmt.Errorf("read analysis failed: %v", err)
		}
		analysis = append(analysis, newAnalysis)
	}

	if len(analysis) == 0 {
		return nil, nil
	}
	return analysis[0], nil
}

func (s *Store) SetAnalysisState(ctx context.Context, analysisId int, isChecked bool) error {
	sql, _, err := goqu.Update("direction_analysis").
		Set(goqu.Record{"is_checked": isChecked}).
		Where(goqu.C("id").Eq(analysisId)).
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

func (s *Store) SetAnalysisFile(ctx context.Context, analysisId int, fileId int) error {
	sql, _, err := goqu.Update("direction_analysis").
		Set(goqu.Record{"file_id": fileId}).
		Where(goqu.C("id").Eq(analysisId)).
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

func readAnalysis(row pgx.Row) (*Analysis, error) {
	var a Analysis

	err := row.Scan(&a.Id, &a.Name, &a.IsChecked, &a.FileId, &a.DirectionId)
	if err != nil {
		return nil, err
	}

	return &a, nil
}

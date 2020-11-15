package store

import (
	"context"
	"fmt"

	"github.com/doug-martin/goqu/v9"
	"github.com/jackc/pgx/v4"
)

type Analysis struct {
	Id        int
	Name      string
	IsChecked bool
}

func (s *Store) GetAnalysisByDirectionId(ctx context.Context, directionId string) ([]*Analysis, error) {
	sql, _, err := goqu.Select("direction_analysis.id", "name", "is_checked").
		From("direction_analysis").
		Where(goqu.C("direction_id").Eq(directionId)).
		LeftJoin(
			goqu.T("analysis"),
			goqu.On(goqu.Ex{
				"analysis.id": goqu.I("direction_analysis.id"),
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

	return analysis, nil
}

func readAnalysis(row pgx.Row) (*Analysis, error) {
	var a Analysis

	err := row.Scan(&a.Id, &a.Name, &a.IsChecked)
	if err != nil {
		return nil, err
	}

	return &a, nil
}
package goqux

import (
	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
)

type InsertOption func(table exp.IdentifierExpression, s *goqu.InsertDataset) *goqu.InsertDataset

func WithInsertDialect(dialect string) InsertOption {
	return func(table exp.IdentifierExpression, s *goqu.InsertDataset) *goqu.InsertDataset {
		return s.WithDialect(dialect)
	}
}

func WithInsertReturningAll() InsertOption {
	return func(table exp.IdentifierExpression, s *goqu.InsertDataset) *goqu.InsertDataset {
		return s.Returning(goqu.Star())
	}
}

func WithInsertReturning(columns ...string) InsertOption {
	return func(table exp.IdentifierExpression, s *goqu.InsertDataset) *goqu.InsertDataset {
		cols := make([]any, 0, len(columns))
		for _, c := range columns {
			cols = append(cols, table.Col(c))
		}
		return s.Returning(cols...)
	}
}

func BuildInsert(tableName string, values []any, options ...InsertOption) (string, []any, error) {
	table := goqu.T(tableName)
	q := goqu.Insert(table).WithDialect(defaultDialect)
	for _, o := range options {
		q = o(table, q)
	}
	encodedValues := make([]map[string]SQLValuer, len(values))
	for i, value := range values {
		encodedValues[i] = encodeValues(value, skipInsert, false)
	}
	return q.Rows(encodedValues).ToSQL()
}

package goqux

import (
	"errors"

	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
)

type UpdateOption func(table exp.IdentifierExpression, s *goqu.UpdateDataset) *goqu.UpdateDataset

func WithUpdateFilters(filters ...goqu.Expression) UpdateOption {
	return func(table exp.IdentifierExpression, s *goqu.UpdateDataset) *goqu.UpdateDataset {
		return s.Where(filters...)
	}
}

func WithUpdateDialect(dialect string) UpdateOption {
	return func(table exp.IdentifierExpression, s *goqu.UpdateDataset) *goqu.UpdateDataset {
		return s.WithDialect(dialect)
	}
}

func WithUpdateReturningAll() UpdateOption {
	return func(table exp.IdentifierExpression, s *goqu.UpdateDataset) *goqu.UpdateDataset {
		return s.Returning(goqu.Star())
	}
}

func WithUpdateReturning(columns ...string) UpdateOption {
	return func(table exp.IdentifierExpression, s *goqu.UpdateDataset) *goqu.UpdateDataset {
		cols := make([]any, 0, len(columns))
		for _, c := range columns {
			cols = append(cols, table.Col(c))
		}
		return s.Returning(cols...)
	}
}

func WithUpdateSet(value any) UpdateOption {
	return func(table exp.IdentifierExpression, s *goqu.UpdateDataset) *goqu.UpdateDataset {
		return s.Set(value)
	}
}

func BuildUpdate(tableName string, value any, options ...UpdateOption) (string, []any, error) {
	table := goqu.T(tableName)
	q := goqu.Update(table).WithDialect(defaultDialect)
	values := encodeValues(value, skipUpdate, true)
	if len(values) == 0 {
		return "", nil, errors.New("no values to update")
	}
	q = q.Set(values)
	for _, o := range options {
		q = o(table, q)
	}
	return q.ToSQL()
}

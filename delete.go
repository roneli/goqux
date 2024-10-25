package goqux

import (
	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
)

type DeleteOption func(table exp.IdentifierExpression, s *goqu.DeleteDataset) *goqu.DeleteDataset

func WithDeleteFilters(filters ...goqu.Expression) DeleteOption {
	return func(table exp.IdentifierExpression, s *goqu.DeleteDataset) *goqu.DeleteDataset {
		return s.Where(filters...)
	}
}

func WithDeleteDialect(dialect string) DeleteOption {
	return func(table exp.IdentifierExpression, s *goqu.DeleteDataset) *goqu.DeleteDataset {
		return s.WithDialect(dialect)
	}
}

func WithDeleteLimit(limit uint) DeleteOption {
	return func(_ exp.IdentifierExpression, s *goqu.DeleteDataset) *goqu.DeleteDataset {
		return s.Limit(limit)
	}
}

func WithDeleteReturningAll() DeleteOption {
	return func(table exp.IdentifierExpression, s *goqu.DeleteDataset) *goqu.DeleteDataset {
		return s.Returning(goqu.Star())
	}
}

func BuildDelete(tableName string, options ...DeleteOption) (string, []any, error) {
	table := goqu.T(tableName)
	deleteQuery := goqu.Delete(table).WithDialect(defaultDialect)
	for _, o := range options {
		deleteQuery = o(table, deleteQuery)
	}
	return deleteQuery.ToSQL()
}

func BuildDeleteDataset(tableName string, options ...DeleteOption) *goqu.DeleteDataset {
	table := goqu.T(tableName)
	deleteQuery := goqu.Delete(table).WithDialect(defaultDialect)
	for _, o := range options {
		deleteQuery = o(table, deleteQuery)
	}
	return deleteQuery
}

package goqux

import (
	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
	"github.com/iancoleman/strcase"
)

type SelectOption func(_ exp.IdentifierExpression, s *goqu.SelectDataset) *goqu.SelectDataset

func WithSelectFilters(filters ...exp.Expression) SelectOption {
	return func(_ exp.IdentifierExpression, s *goqu.SelectDataset) *goqu.SelectDataset {
		return s.Where(filters...)
	}
}

func WithSelectDialect(dialect string) SelectOption {
	return func(_ exp.IdentifierExpression, s *goqu.SelectDataset) *goqu.SelectDataset {
		return s.WithDialect(dialect)
	}
}

func WithSelectLimit(limit uint) SelectOption {
	return func(_ exp.IdentifierExpression, s *goqu.SelectDataset) *goqu.SelectDataset {
		return s.Limit(limit)
	}
}

func WithSelectOffset(offset uint) SelectOption {
	return func(_ exp.IdentifierExpression, s *goqu.SelectDataset) *goqu.SelectDataset {
		return s.Offset(offset)
	}
}

func WithKeySet(columns []string, values []any) SelectOption {
	return func(table exp.IdentifierExpression, s *goqu.SelectDataset) *goqu.SelectDataset {
		if values == nil {
			return s
		}
		for i, c := range columns {
			s = s.Where(table.Col(strcase.ToSnake(c)).Gt(values[i])).Order(table.Col(strcase.ToSnake(c)).Asc())
		}
		// Make sure to clear offset with KeySet pagination
		return s.ClearOffset()
	}
}

func WithSelectOrder(order ...exp.OrderedExpression) SelectOption {
	return func(_ exp.IdentifierExpression, s *goqu.SelectDataset) *goqu.SelectDataset {
		return s.Order(order...)
	}
}

func WithSelectStar() SelectOption {
	return func(_ exp.IdentifierExpression, s *goqu.SelectDataset) *goqu.SelectDataset {
		return s.Select(goqu.Star())
	}
}

func BuildSelect[T any](tableName string, dst T, options ...SelectOption) (string, []any, error) {
	table := goqu.T(tableName)
	selectQuery := goqu.Select(getColumnsFromStruct(table, dst, skipSelect)...).From(table).WithDialect(defaultDialect)
	for _, o := range options {
		selectQuery = o(table, selectQuery)
	}
	return selectQuery.ToSQL()
}

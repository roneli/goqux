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
			for _, c := range columns {
				s = s.Order(table.Col(strcase.ToSnake(c)).Asc())
			}
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

type JoinOp struct {
	Table string
	On    exp.JoinCondition
}

// WithInnerJoinSelection returns a select option that inner joins the given table on the given column by tableName.column = otherTable.otherColumn,
// as well as selecting the columns from the given struct. each top-level struct field will be treated as a table and each field within that struct
// will be treated as a column.
func WithInnerJoinSelection[T any](op ...JoinOp) SelectOption {
	return func(_ exp.IdentifierExpression, s *goqu.SelectDataset) *goqu.SelectDataset {
		for _, j := range op {
			s = s.InnerJoin(goqu.T(j.Table), j.On)
		}
		selectFields := make([]any, 0)
		for _, c := range getSelectionFieldsFromSelectionStruct(new(T)) {
			selectFields = append(selectFields, c)
		}
		return s.Select(selectFields...)
	}
}

// WithLeftJoinSelection returns a select option that left joins the given table on the given column by tableName.column = otherTable.otherColumn,
// as well as selecting the columns from the given struct. each top-level struct field will be treated as a table and each field within that struct
// will be treated as a column.
func WithLeftJoinSelection[T any](op ...JoinOp) SelectOption {
	return func(_ exp.IdentifierExpression, s *goqu.SelectDataset) *goqu.SelectDataset {
		for _, j := range op {
			s = s.LeftJoin(goqu.T(j.Table), j.On)
		}
		selectFields := make([]any, 0)
		for _, c := range getSelectionFieldsFromSelectionStruct(new(T)) {
			selectFields = append(selectFields, c)
		}
		return s.Select(selectFields...)
	}
}

func BuildSelect[T any](tableName string, dst T, options ...SelectOption) (string, []any, error) {
	table := goqu.T(tableName)
	structCols := make([]any, 0)
	for _, c := range getColumnsFromStruct(table, dst, skipSelect) {
		structCols = append(structCols, c)
	}
	selectQuery := goqu.Dialect(defaultDialect).Select(structCols...).From(table)
	for _, o := range options {
		selectQuery = o(table, selectQuery)
	}
	return selectQuery.ToSQL()
}

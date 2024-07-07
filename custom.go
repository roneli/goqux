package goqux

import (
	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
)

type CustomColumn interface {
	BuildSelect(table exp.IdentifierExpression, col string) goqu.Expression
	BuildInsert(value any) goqu.Expression
}

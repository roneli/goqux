package goqux_test

import (
	"testing"

	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
	"github.com/roneli/goqux"
	"github.com/stretchr/testify/assert"
)

type testCustomString string

func (t testCustomString) BuildSelect(table exp.IdentifierExpression, col string) goqu.Expression {
	return goqu.Func("lower", table.Col("test_column")).As(table.Col("test_column"))
}

func (t testCustomString) BuildInsert(value any) goqu.Expression {
	return goqu.Func("upper", value)
}

type customModel struct {
	TestColumn testCustomString `db:"test_column"`
}

type customJoinModel struct {
	T2     customModel `db:"custom_models"`
	Table1 *Table1
}

func Test_BuildSelectWithCustomColumn(t *testing.T) {
	tableTests := []struct {
		name          string
		dst           interface{}
		expectedQuery string
		options       []goqux.SelectOption
		expectedArgs  []interface{}
		expectedError error
	}{
		{
			name:          "simple select",
			dst:           customModel{},
			expectedQuery: `SELECT lower("custom_models"."test_column") AS "custom_models"."test_column" FROM "custom_models"`,
		},
		{
			name: "select with inner join",
			dst:  customJoinModel{},
			options: []goqux.SelectOption{goqux.WithInnerJoinSelection[customJoinModel](goqux.JoinOp{
				Table: "table_1",
				On:    goqu.On(goqux.Column("table_1", "string_field").Eq(goqux.Column("custom_models", "test_column"))),
			})},
			expectedQuery: `SELECT lower("custom_models"."test_column") AS "custom_models"."test_column", "table_1"."int_field" AS "table_1"."int_field", "table_1"."cool_field" AS "table_1"."cool_field" FROM "custom_models" INNER JOIN "table_1" ON ("table_1"."string_field" = "custom_models"."test_column")`,
		},
	}
	for _, tt := range tableTests {
		t.Run(tt.name, func(t *testing.T) {
			query, args, err := goqux.BuildSelect("custom_models", tt.dst, tt.options...)
			if tt.expectedError != nil {
				assert.ErrorIs(t, tt.expectedError, err)
			}
			assert.Equal(t, tt.expectedQuery, query)
			assert.ElementsMatch(t, tt.expectedArgs, args)
		})
	}
}

func Test_BuildInsertWithCustomColumn(t *testing.T) {
	{
		testTables := []struct {
			name          string
			values        []any
			opts          []goqux.InsertOption
			expectedQuery string
			expectedArgs  []interface{}
			expectedError error
		}{
			{
				name:          "simple_insert",
				values:        []any{customModel{TestColumn: "test-val"}},
				expectedQuery: `INSERT INTO "custom_models" ("test_column") VALUES ($1)`,
				expectedArgs:  []interface{}{""},
				expectedError: nil,
			},
		}
		for _, tt := range testTables {
			t.Run(tt.name, func(t *testing.T) {
				query, args, err := goqux.BuildInsert("custom_models", tt.values, tt.opts...)
				if tt.expectedError != nil {
					assert.ErrorIs(t, tt.expectedError, err)
				}
				assert.Equal(t, tt.expectedQuery, query)
				assert.Equal(t, tt.expectedArgs, args)
			})
		}
	}
}

package goqux_test

import (
	"testing"

	"github.com/doug-martin/goqu/v9"
	"github.com/roneli/goqux"
	"github.com/stretchr/testify/assert"
)

type selectModel struct {
	IntField       int
	invisibleField bool
	fieldToSkip    int `goqux:"skip_select"`
}

type joinModel struct {
	T2     selectModel `db:"select_models"`
	Table1 *Table1
}

type Table1 struct {
	IntField    int
	StringField string `db:"cool_field"`
}

type doubleJoinModel struct {
	T2     selectModel `db:"select_models"`
	Table1 *Table1
	T3     Table2 `db:"table_2"`
}

type Table2 struct {
	IntField    int
	StringField string `db:"cool_field"`
}

func TestBuildSelect(t *testing.T) {
	tableTests := []struct {
		name          string
		dst           interface{}
		options       []goqux.SelectOption
		expectedQuery string
		expectedArgs  []interface{}
		expectedError error
	}{
		{
			name:          "simple_select",
			dst:           selectModel{},
			expectedQuery: `SELECT "select_models"."int_field" FROM "select_models"`,
			expectedArgs:  []interface{}{},
			expectedError: nil,
		},
		{
			name:          "select_with_filters",
			dst:           selectModel{},
			options:       []goqux.SelectOption{goqux.WithSelectFilters(goqux.Column("select_models", "int_field").Eq(1))},
			expectedQuery: `SELECT "select_models"."int_field" FROM "select_models" WHERE ("select_models"."int_field" = $1)`,
			expectedArgs:  []interface{}{int64(1)},
		},
		{
			name:          "select_with_limit",
			dst:           selectModel{},
			options:       []goqux.SelectOption{goqux.WithSelectLimit(1)},
			expectedQuery: `SELECT "select_models"."int_field" FROM "select_models" LIMIT $1`,
			expectedArgs:  []interface{}{int64(1)},
		},
		{
			name:          "select_with_offset",
			dst:           selectModel{},
			options:       []goqux.SelectOption{goqux.WithSelectOffset(1)},
			expectedQuery: `SELECT "select_models"."int_field" FROM "select_models" OFFSET $1`,
			expectedArgs:  []interface{}{int64(1)},
		},
		{
			name:          "select_with_dialect",
			dst:           selectModel{},
			options:       []goqux.SelectOption{goqux.WithSelectDialect("postgres"), goqux.WithSelectOffset(1)},
			expectedQuery: `SELECT "select_models"."int_field" FROM "select_models" OFFSET $1`,
			expectedArgs:  []interface{}{int64(1)},
		},
		{
			name:          "select_with_order",
			dst:           selectModel{},
			options:       []goqux.SelectOption{goqux.WithSelectOrder(goqux.Column("select_models", "int_field").Desc())},
			expectedQuery: `SELECT "select_models"."int_field" FROM "select_models" ORDER BY "select_models"."int_field" DESC`,
			expectedArgs:  []interface{}{},
		},
		{
			name: "select_with_inner_join_selection",
			dst:  joinModel{},
			options: []goqux.SelectOption{goqux.WithInnerJoinSelection[joinModel](goqux.JoinOp{
				Table: "table_1",
				On:    goqu.On(goqux.Column("table_1", "int_field").Eq(goqux.Column("select_models", "int_field"))),
			})},
			expectedQuery: `SELECT "select_models"."int_field" AS "select_models.int_field", "table_1"."int_field" AS "table_1.int_field", "table_1"."cool_field" AS "table_1.cool_field" FROM "select_models" INNER JOIN "table_1" ON ("table_1"."int_field" = "select_models"."int_field")`,
		},
		{
			name: "select_with_left_selection",
			dst:  joinModel{},
			options: []goqux.SelectOption{goqux.WithLeftJoinSelection[joinModel](goqux.JoinOp{
				Table: "table_1",
				On:    goqu.On(goqux.Column("table_1", "int_field").Eq(goqux.Column("select_models", "int_field"))),
			})},
			expectedQuery: `SELECT "select_models"."int_field" AS "select_models.int_field", "table_1"."int_field" AS "table_1.int_field", "table_1"."cool_field" AS "table_1.cool_field" FROM "select_models" LEFT JOIN "table_1" ON ("table_1"."int_field" = "select_models"."int_field")`,
		},
		{
			name: "select_with_double_join_selection",
			dst:  doubleJoinModel{},
			options: []goqux.SelectOption{goqux.WithInnerJoinSelection[doubleJoinModel](goqux.JoinOp{
				Table: "table_1",
				On:    goqu.On(goqux.Column("table_1", "int_field").Eq(goqux.Column("select_models", "int_field"))),
			}, goqux.JoinOp{
				Table: "table_2",
				On:    goqu.On(goqux.Column("table_2", "int_field").Eq(goqux.Column("select_models", "int_field"))),
			})},
			expectedQuery: `SELECT "select_models"."int_field" AS "select_models.int_field", "table_1"."int_field" AS "table_1.int_field", "table_1"."cool_field" AS "table_1.cool_field", "table_2"."int_field" AS "table_2.int_field", "table_2"."cool_field" AS "table_2.cool_field" FROM "select_models" INNER JOIN "table_1" ON ("table_1"."int_field" = "select_models"."int_field") INNER JOIN "table_2" ON ("table_2"."int_field" = "select_models"."int_field")`,
		},
	}
	for _, tableTest := range tableTests {
		t.Run(tableTest.name, func(t *testing.T) {
			query, args, err := goqux.BuildSelect("select_models", tableTest.dst, tableTest.options...)
			if tableTest.expectedError != nil {
				assert.ErrorIs(t, tableTest.expectedError, err)
			}
			assert.Equal(t, tableTest.expectedQuery, query)
			assert.ElementsMatch(t, tableTest.expectedArgs, args)
		})
	}
}

package goqux_test

import (
	"testing"

	"github.com/roneli/goqux"
	"github.com/stretchr/testify/assert"
)

type selectModel struct {
	IntField       int
	invisibleField bool
	fieldToSkip    int `goqux:"skip_select"`
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

package goqux_test

import (
	"errors"
	"testing"

	"github.com/roneli/goqux"
	"github.com/stretchr/testify/assert"
)

type updateModel struct {
	IntField       int
	DbTag          string `db:"another_col_name"`
	DbTagOmitEmpty string `db:"another_col_name_omit,omitempty"`
}

func TestBuildUpdate(t *testing.T) {
	tableTests := []struct {
		name          string
		dst           interface{}
		options       []goqux.UpdateOption
		expectedQuery string
		expectedArgs  []interface{}
		expectedError error
	}{
		{
			name:          "simple_update",
			dst:           updateModel{IntField: 5, DbTag: "test"},
			expectedQuery: `UPDATE "update_models" SET "another_col_name"=$1,"int_field"=$2`,
			expectedArgs:  []interface{}{"test", int64(5)},
			expectedError: nil,
		},
		{
			name:          "simple_update",
			dst:           updateModel{IntField: 5, DbTagOmitEmpty: "test"},
			expectedQuery: `UPDATE "update_models" SET "another_col_name_omit"=$1,"int_field"=$2`,
			expectedArgs:  []interface{}{"test", int64(5)},
			expectedError: nil,
		},
		{
			name:          "update_with_filters",
			dst:           updateModel{IntField: 5},
			options:       []goqux.UpdateOption{goqux.WithUpdateFilters(goqux.Column("update_models", "int_field").Eq(1))},
			expectedQuery: `UPDATE "update_models" SET "int_field"=$1 WHERE ("update_models"."int_field" = $2)`,
			expectedArgs:  []interface{}{int64(5), int64(1)},
		},
		{
			name:          "update_with_returning",
			dst:           updateModel{IntField: 5},
			options:       []goqux.UpdateOption{goqux.WithUpdateReturningAll()},
			expectedQuery: `UPDATE "update_models" SET "int_field"=$1 RETURNING *`,
			expectedArgs:  []interface{}{int64(5)},
		},
		{
			name:          "update_with_zero_values",
			dst:           updateModel{IntField: 0},
			expectedQuery: ``,
			expectedArgs:  nil,
			expectedError: errors.New("no values to update"),
		},
	}
	for _, tt := range tableTests {
		t.Run(tt.name, func(t *testing.T) {
			query, args, err := goqux.BuildUpdate("update_models", tt.dst, tt.options...)
			if tt.expectedError != nil {
				assert.Equal(t, tt.expectedError, err)
			}
			assert.Equal(t, tt.expectedQuery, query)
			assert.Equal(t, tt.expectedArgs, args)
		})
	}
}

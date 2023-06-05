package goqux_test

import (
	"testing"

	"github.com/roneli/goqux"
	"github.com/stretchr/testify/assert"
)

type deleteModel struct {
	IntField       int
	invisibleField bool
	FieldToSkip    int `goqux:"skip_delete"`
}

func TestBuildDelete(t *testing.T) {
	tableTests := []struct {
		name          string
		dst           interface{}
		options       []goqux.DeleteOption
		expectedQuery string
		expectedArgs  []interface{}
		expectedError error
	}{
		{
			name:          "simple_delete",
			dst:           deleteModel{},
			expectedQuery: `DELETE FROM "delete_models"`,
			expectedArgs:  []interface{}{},
			expectedError: nil,
		},
		{
			name:          "delete_with_filters",
			dst:           deleteModel{},
			options:       []goqux.DeleteOption{goqux.WithDeleteFilters(goqux.Column("delete_models", "int_field").Eq(1))},
			expectedQuery: `DELETE FROM "delete_models" WHERE ("delete_models"."int_field" = $1)`,
			expectedArgs:  []interface{}{int64(1)},
		},
		{
			name:          "delete_with_returning",
			dst:           deleteModel{},
			options:       []goqux.DeleteOption{goqux.WithDeleteReturningAll()},
			expectedQuery: `DELETE FROM "delete_models" RETURNING *`,
			expectedArgs:  []interface{}{},
		},
		{
			name:          "delete_with_dialect",
			dst:           deleteModel{},
			options:       []goqux.DeleteOption{goqux.WithDeleteDialect("postgres"), goqux.WithDeleteFilters(goqux.Column("delete_models", "int_field").Eq(1))},
			expectedQuery: `DELETE FROM "delete_models" WHERE ("delete_models"."int_field" = $1)`,
			expectedArgs:  []interface{}{int64(1)},
		},
	}
	for _, tt := range tableTests {
		t.Run(tt.name, func(t *testing.T) {
			query, args, err := goqux.BuildDelete("delete_models", tt.options...)
			if tt.expectedError != nil {
				assert.ErrorIs(t, tt.expectedError, err)
			}
			assert.Equal(t, tt.expectedQuery, query)
			assert.ElementsMatch(t, tt.expectedArgs, args)
		})
	}
}

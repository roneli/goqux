package goqux_test

import (
	"testing"

	"github.com/roneli/goqux"
	"github.com/stretchr/testify/assert"
)

type insertModel struct {
	IntField       int64 `db:"int_field"`
	OtherValue     string
	SkipInsert     bool   `goqux:"skip_insert"`
	DbTag          string `db:"another_col_name"`
	DbTagOmitEmpty string `db:"another_col_name_omit,omitempty"`
}

func TestBuildInsert(t *testing.T) {
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
			values:        []any{insertModel{IntField: 5, DbTagOmitEmpty: "test"}},
			expectedQuery: `INSERT INTO "insert_models" ("another_col_name", "another_col_name_omit", "int_field", "other_value") VALUES ($1, $2, $3, $4)`,
			expectedArgs:  []interface{}{"", "test", int64(5), ""},
			expectedError: nil,
		},
		{
			name:          "simple_insert_ompitempty",
			values:        []any{insertModel{IntField: 5}},
			expectedQuery: `INSERT INTO "insert_models" ("another_col_name", "int_field", "other_value") VALUES ($1, $2, $3)`,
			expectedArgs:  []interface{}{"", int64(5), ""},
			expectedError: nil,
		},
		{
			name: "insert_multiple_values",
			values: []any{
				insertModel{IntField: 5},
				insertModel{IntField: 6},
			},
			expectedQuery: `INSERT INTO "insert_models" ("another_col_name", "int_field", "other_value") VALUES ($1, $2, $3), ($4, $5, $6)`,
			expectedArgs:  []interface{}{"", int64(5), "", "", int64(6), ""},
			expectedError: nil,
		},
		{
			name: "insert_with_returning_all",
			values: []any{
				insertModel{IntField: 5},
			},
			opts:          []goqux.InsertOption{goqux.WithInsertReturningAll()},
			expectedQuery: `INSERT INTO "insert_models" ("another_col_name", "int_field", "other_value") VALUES ($1, $2, $3) RETURNING *`,
			expectedArgs:  []interface{}{"", int64(5), ""},
		},
		{
			name: "insert_with_not_prepared",
			values: []any{
				insertModel{IntField: 5},
				insertModel{IntField: 6},
			},
			opts:          []goqux.InsertOption{goqux.WithNotPrepend()},
			expectedQuery: `INSERT INTO "insert_models" ("another_col_name", "int_field", "other_value") VALUES ('', 5, ''), ('', 6, '')`,
			expectedArgs:  []interface{}{},
		},
	}
	for _, tt := range testTables {
		t.Run(tt.name, func(t *testing.T) {
			query, args, err := goqux.BuildInsert("insert_models", tt.values, tt.opts...)
			if tt.expectedError != nil {
				assert.ErrorIs(t, tt.expectedError, err)
			}
			assert.Equal(t, tt.expectedQuery, query)
			assert.Equal(t, tt.expectedArgs, args)
		})
	}
}

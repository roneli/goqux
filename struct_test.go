package goqux

import (
	"testing"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncodeValues(t *testing.T) {
	tableTests := []struct {
		name           string
		model          interface{}
		values         map[string]SQLValuer
		skipFlag       string
		skipZeroValues bool
	}{
		{
			name: "encode_insert",
			model: struct {
				IntField    int
				unexported  bool
				FieldToSkip int `goqux:"skip_insert"`
			}{IntField: 1},
			values:         map[string]SQLValuer{"int_field": {1}},
			skipFlag:       skipInsert,
			skipZeroValues: false,
		},
		{
			name: "encode_array",
			model: struct {
				IntField []int
			}{IntField: []int{1, 2}},
			values:         map[string]SQLValuer{"int_field": {[]int{1, 2}}},
			skipFlag:       skipInsert,
			skipZeroValues: false,
		},
		{
			name: "encode_array_with_zero_values",
			model: struct {
				IntField []int
			}{IntField: []int{1, 0}},
			values:         map[string]SQLValuer{"int_field": {[]int{1, 0}}},
			skipFlag:       skipInsert,
			skipZeroValues: false,
		},
		{
			name: "encode_zero_values",
			model: struct {
				IntField    int
				FloatField  float64
				unexported  bool
				FieldToSkip int `goqux:"skip_insert"`
			}{IntField: 0},
			values:         map[string]SQLValuer{},
			skipFlag:       skipInsert,
			skipZeroValues: true,
		},
		{
			name: "encode_map_values",
			model: struct {
				MapValue map[string]any `goqux:"skip_compare"`
			}{MapValue: map[string]any{"type": "map"}},
			values:         map[string]SQLValuer{"map_value": {map[string]any{"type": "map"}}},
			skipFlag:       skipInsert,
			skipZeroValues: false,
		},
		{
			name: "encode_empty_struct",
			model: struct {
				IntField    int
				FloatField  float64
				unexported  bool
				EmptyStruct struct {
					Type  string
					Value string
				}
			}{},
			values:         map[string]SQLValuer{},
			skipFlag:       skipInsert,
			skipZeroValues: true,
		},
		{
			name: "encode_omitempty_nonempty",
			model: struct {
				IntField int `db:"int_field,omitempty"`
			}{IntField: 3},
			values:         map[string]SQLValuer{"int_field": {3}},
			skipFlag:       skipInsert,
			skipZeroValues: false,
		},
		{
			name: "encode_omitempty_empty",
			model: struct {
				IntField int `db:"int_field,omitempty"`
			}{IntField: 0},
			values:         map[string]SQLValuer{},
			skipFlag:       skipInsert,
			skipZeroValues: false,
		},
	}
	for _, tt := range tableTests {
		t.Run(tt.name, func(t *testing.T) {
			values := encodeValues(tt.model, tt.skipFlag, tt.skipZeroValues)
			assert.Equal(t, tt.values, values)
		})
	}
}

func TestEncodeTimeValue(t *testing.T) {
	values := encodeValues(struct {
		TimeField   *time.Time `goqux:"now"`
		unexported  bool
		FieldToSkip int `goqux:"skip_insert"`
	}{
		FieldToSkip: 5,
	}, skipInsert, true)
	if tf, ok := values["time_field"]; ok {
		require.NotNil(t, tf)
		return
	}
}

func TestGetColumnsFromStruct(t *testing.T) {
	tableTests := []struct {
		name     string
		model    interface{}
		expected []exp.IdentifierExpression
	}{
		{
			name: "get_columns_from_struct",
			model: struct {
				IntField    int
				unexported  bool
				FieldToSkip int `goqux:"skip_select"`
				DbOsField   int `db:"db_field"`
			}{},
			expected: []exp.IdentifierExpression{goqu.T("table").Col("int_field"), goqu.T("table").Col("db_field")},
		},
		{
			name: "get_columns_from_struct_with_omitempty",
			model: struct {
				IntField    int `db:"int_field,omitempty"` // should not be skipped
				StringField string
			}{},
			expected: []exp.IdentifierExpression{goqu.T("table").Col("int_field"), goqu.T("table").Col("string_field")},
		},
	}
	for _, tt := range tableTests {
		t.Run(tt.name, func(t *testing.T) {
			columns := getColumnsFromStruct(goqu.T("table"), tt.model, skipSelect)
			assert.Equal(t, tt.expected, columns)
		})
	}
}

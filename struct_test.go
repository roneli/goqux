package goqux

import (
	"testing"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEncodeValues(t *testing.T) {
	tableTests := []struct {
		name     string
		model    interface{}
		values   map[string]SQLValuer
		skipFlag string
	}{
		{
			name: "encode_insert",
			model: struct {
				IntField    int
				unexported  bool
				FieldToSkip int `goqux:"skip_insert"`
			}{IntField: 1},
			values:   map[string]SQLValuer{"int_field": {1}},
			skipFlag: skipInsert,
		},
		{
			name: "encode_array",
			model: struct {
				IntField []int
			}{IntField: []int{1, 2}},
			values:   map[string]SQLValuer{"int_field": {[]int{1, 2}}},
			skipFlag: skipInsert,
		},
	}
	for _, tt := range tableTests {
		t.Run(tt.name, func(t *testing.T) {
			values := encodeValues(tt.model, tt.skipFlag)
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
	}, skipInsert)
	if tf, ok := values["time_field"]; ok {
		require.NotNil(t, tf)
		return
	}
	t.Fail()
}

func TestGetColumnsFromStruct(t *testing.T) {
	type selectModel struct {
		IntField    int
		unexported  bool
		FieldToSkip int `goqux:"skip_select"`
	}
	columns := getColumnsFromStruct(goqu.T("table"), selectModel{}, skipSelect)
	assert.Equal(t, []interface{}{goqu.T("table").Col("int_field")}, columns)
}

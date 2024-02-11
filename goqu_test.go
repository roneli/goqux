package goqux

import (
	"database/sql/driver"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

type structField struct {
	V map[string]interface{} `json:"v"`
}

func (s structField) Value() (driver.Value, error) {
	return json.Marshal(s)
}

type structValuerArray []structField

func (s structValuerArray) Value() (driver.Value, error) {
	return json.Marshal(s)
}

func TestSQLValuer_Value(t *testing.T) {
	tableTestValues := []struct {
		name          string
		value         any
		expected      interface{}
		expectedError error
	}{
		{
			name:          "string",
			value:         "test",
			expected:      "test",
			expectedError: nil,
		},
		{
			name:          "int",
			value:         1,
			expected:      1,
			expectedError: nil,
		},
		{
			name: "map[string]interface{}",
			value: map[string]interface{}{
				"test": "test",
			},
			expected:      []byte(`{"test":"test"}`),
			expectedError: nil,
		},
		{
			name: "map[string]string",
			value: map[string]string{
				"test": "test",
			},
			expected: []byte(`{"test":"test"}`),
		},
		{
			name: "[]map[string]interface{}",
			value: []map[string]interface{}{
				{
					"test": "test",
				}, {
					"test": "test2",
				},
			},
			expected: []byte(`[{"test":"test"},{"test":"test2"}]`),
		},
		{
			name: "[]interface{}",
			value: []interface{}{
				"test",
			},
			expected: []byte(`["test"]`),
		},
		{
			name: "[]string",
			value: []string{
				"test",
			},
			expected: "{\"test\"}",
		},
		{
			name: "struct_implements_valuer",
			value: structField{
				V: map[string]interface{}{
					"test": "test",
				},
			},
			expected: []byte(`{"v":{"test":"test"}}`),
		},
		{
			name: "valuer_pointer",
			value: &structField{
				V: map[string]interface{}{
					"test": "test",
				},
			},
			expected: []byte(`{"v":{"test":"test"}}`),
		},
		{
			name:     "valuer_nil_pointer",
			value:    (*structField)(nil),
			expected: nil,
		},
		{
			name:     "empty_struct",
			value:    structField{},
			expected: []byte(`{"v":null}`),
		},
		{
			name:     "empty_struct_pointer",
			value:    &structField{},
			expected: []byte(`{"v":null}`),
		},
		{
			name: "valuer_array_struct",
			value: structValuerArray{
				{
					V: map[string]interface{}{
						"test": "test",
					},
				},
			},
			expected: []byte(`[{"v":{"test":"test"}}]`),
		},
		{
			name: "valuer_array_pointer",
			value: &structValuerArray{
				{
					V: map[string]interface{}{
						"test": "test",
					},
				},
			},
			expected: []byte(`[{"v":{"test":"test"}}]`),
		},
	}
	for _, tt := range tableTestValues {
		t.Run(tt.name, func(t *testing.T) {
			sv := SQLValuer{tt.value}
			got, err := sv.Value()
			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.Equal(t, tt.expected, got)
			}
		})
	}
}

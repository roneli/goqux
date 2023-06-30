package goqux

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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

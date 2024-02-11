package goqux

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
	"github.com/google/uuid"
	"github.com/iancoleman/strcase"
	"github.com/lib/pq"
)

var defaultDialect = "postgres"

func init() {
	goqu.SetColumnRenameFunction(strcase.ToSnake)
	goqu.SetDefaultPrepared(true)
}

// Column is shorthand for goqu.T(table).Col(column).
func Column(table string, column string) exp.IdentifierExpression {
	return goqu.T(table).Col(column)
}

// SetDefaultDialect sets the default dialect for goqux.
func SetDefaultDialect(dialect string) {
	defaultDialect = dialect
}

// SQLValuer is the valuer struct that is used for goqu rows conversion.
type SQLValuer struct {
	V interface{}
}

// Value converts the given value to the correct drive.Value.
func (t SQLValuer) Value() (driver.Value, error) {
	switch t.V.(type) {
	case []string, []bool, []float32, []float64, []int, []int64, []int32:
		value, err := pq.Array(t.V).Value()
		if err != nil {
			return nil, fmt.Errorf("failed to convert value: %w", err)
		}
		return value, nil
	case map[string]interface{}, []map[string]interface{}, []interface{}, map[string]string:
		return json.Marshal(t.V)
	case uuid.UUID:
		if t.V == uuid.Nil {
			return nil, nil
		}
		return t.V, nil
	default:
		// if we didn't find a common type use reflection to guess if the type is of map
		if reflect.TypeOf(t.V).Kind() == reflect.Map {
			return json.Marshal(t.V)
		} else if reflect.TypeOf(t.V).Kind() == reflect.Pointer {
			if reflect.ValueOf(t.V).IsZero() {
				return nil, nil
			}
			s := reflect.ValueOf(t.V).Elem().Interface()
			if valuer, ok := s.(driver.Valuer); ok {
				return valuer.Value()
			}
		}
		if valuer, ok := t.V.(driver.Valuer); ok {
			return valuer.Value()
		}
		return t.V, nil
	}
}

package goqux

import (
	"reflect"
	"strings"
	"time"

	"fmt"
	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	"github.com/doug-martin/goqu/v9/exp"
	"github.com/iancoleman/strcase"
)

const (
	// TagName goqux tag field query building
	tagName = "goqux"
	// TagName of db to change column field
	tagNameDb = "db"
	// skip is field that allows skipping the column
	skipSelect          = "skip_select"
	skipUpdate          = "skip_update"
	skipInsert          = "skip_insert"
	skipReturningDelete = "skip_delete"
	// if the field is of type time.Time it will inject time.Now
	defaultNow = "now"
	// Same as default now but will inject time.Now().UTC()
	defaultNowUtc = "now_utc"
	// omitempty will skip the field if it is zero value
	omitEmpty = ",omitempty"
)

func convertMapToSQLValuer(m map[string]any) map[string]SQLValuer {
	values := make(map[string]SQLValuer)
	for k, v := range m {
		values[k] = SQLValuer{v}
	}
	return values
}

func encodeValues(v any, skipType string, skipZeroValues bool) map[string]SQLValuer {
	t := reflect.ValueOf(v)
	// if we received a map we will just convert it to a map of SQLValuer
	if t.Kind() == reflect.Map {
		return convertMapToSQLValuer(v.(map[string]any))
	}

	// if we received a pointer we will just derference it
	if t.Kind() == reflect.Pointer {
		t = t.Elem() // Dereference to get the underlying type
	}

	fields := reflect.VisibleFields(t.Type())
	values := make(map[string]SQLValuer)
	for _, f := range fields {
		if !f.IsExported() || strings.Contains(f.Tag.Get(tagName), skipType) {
			continue
		}
		value := t.FieldByName(f.Name)
		// We want to support the case when there is no value in one of the fields
		if skipZeroValues && value.IsZero() {
			continue
		}

		columnName := strcase.ToSnake(f.Name)
		if dbTag := f.Tag.Get(tagNameDb); dbTag != "" {
			if strings.Contains(dbTag, omitEmpty) {
				if value.IsZero() {
					continue
				}
				dbTag = cleanDbTag(dbTag)
			}
			columnName = dbTag
		}

		switch {
		case strings.Contains(f.Tag.Get(tagName), defaultNowUtc):
			values[columnName] = SQLValuer{time.Now().UTC()}
		case strings.Contains(f.Tag.Get(tagName), defaultNow):
			values[columnName] = SQLValuer{time.Now()}
		default:
			values[columnName] = SQLValuer{value.Interface()}
		}
	}
	return values
}

func getColumnsFromStruct(table exp.IdentifierExpression, s any, skipType string) []exp.IdentifierExpression {
	t := reflect.TypeOf(s)
	// if we received a pointer we will just derference it
	if t.Kind() == reflect.Pointer {
		t = t.Elem() // Dereference to get the underlying type
	}
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	fields := reflect.VisibleFields(t)
	var cols = make([]exp.IdentifierExpression, 0)
	for _, f := range fields {
		if !f.IsExported() || strings.Contains(f.Tag.Get(tagName), skipType) {
			continue
		}
		var colName string
		if dbTag := f.Tag.Get(tagNameDb); dbTag != "" {
			colName = cleanDbTag(dbTag)
		} else {
			colName = strcase.ToSnake(f.Name)
		}
		cols = append(cols, table.Col(colName))
	}
	return cols
}

func cleanDbTag(tag string) string {
	if strings.Contains(tag, omitEmpty) {
		tag = strings.ReplaceAll(tag, omitEmpty, "")
	}

	return tag
}

func getSelectionFieldsFromSelectionStruct(s interface{}) []exp.AliasedExpression {
	cols := make([]exp.AliasedExpression, 0)
	t := reflect.TypeOf(s)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	tableFields := reflect.VisibleFields(t)
	for _, tf := range tableFields {
		if !tf.IsExported() {
			continue
		}
		if tf.Type.Kind() != reflect.Struct && !(tf.Type.Kind() == reflect.Ptr && tf.Type.Elem().Kind() == reflect.Struct) {
			continue
		}
		tableName := strcase.ToSnake(tf.Name)
		if dbTag := tf.Tag.Get(tagNameDb); dbTag != "" {
			tableName = cleanDbTag(dbTag)
		}
		subTableColumns := getColumnsFromStruct(goqu.T(tableName), reflect.New(tf.Type).Elem().Interface(), skipSelect)
		for _, c := range subTableColumns {
			// SELECT "table"."column" AS "table.column" will make sure dbscan scans all the columns correctly
			cc := c.GetCol()
			cName := tableName + "." + cc.(string)
			cols = append(cols, goqu.T(tableName).Col(cc).As(goqu.C(cName)))
		}
	}
	return cols
}

func keysetColumns[T any](columns []string, strct T) ([]string, error) {
	var cols []string
	t := reflect.TypeOf(strct)

	// If T is a pointer, dereference it safely
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// Ensure `strct` is a struct
	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("T must be a struct type")
	}

	// Process each column specified
	for _, c := range columns {
		found := false
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			tagValue := field.Tag.Get("db")

			// Check if the field name or "db" tag matches
			if field.Name == c || tagValue == c {
				found = true
				// Always prefer a set db tag value
				if tagValue != "" {
					cols = append(cols, tagValue)
				} else {
					cols = append(cols, strcase.ToSnake(field.Name))
				}
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("column %s not found in struct %s", c, t.Name())
		}
	}

	return cols, nil
}

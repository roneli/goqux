package goqux

import (
	"reflect"
	"strings"
	"time"

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

type colExpression struct {
	col        string
	expression goqu.Expression
}

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

		custom, ok := value.Interface().(CustomColumn)
		if ok {
			values[columnName] = SQLValuer{custom.BuildInsert(value)}
		}
	}
	return values
}

func getColumnsFromStruct(table exp.IdentifierExpression, s any, skipType string) []colExpression {
	t := reflect.TypeOf(s)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	fields := reflect.VisibleFields(t)
	var cols = make([]colExpression, 0)
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

		// Get the interface type
		customColumnType := reflect.TypeOf((*CustomColumn)(nil)).Elem()
		// Check if the type implements the interface
		var expression goqu.Expression
		if reflect.PtrTo(f.Type).Implements(customColumnType) {
			// Create a new instance of the field type and call MyMethod
			customColumn := reflect.New(f.Type).Interface().(CustomColumn)
			expression = customColumn.BuildSelect(table, colName)
		} else {
			expression = table.Col(colName)
		}
		cols = append(cols, colExpression{col: colName, expression: expression})

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
			cName := tableName + "." + c.col
			aliased, ok := c.expression.(exp.AliasedExpression)
			if ok {
				cols = append(cols, aliased)
			} else {
				cols = append(cols, exp.NewAliasExpression(c.expression, cName))
			}
		}
	}
	return cols
}

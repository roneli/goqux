package goqux

import (
	"fmt"
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
	omitEmpty = "omitempty"
	// omitnil will skip the field if it is nil
	omitNil = "omitnil"
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
				if !value.IsValid() || value.IsZero() { // IsZero panic on valid values
					continue
				}
				dbTag = cleanDbTag(dbTag, omitEmpty)
			}
			if strings.Contains(dbTag, omitNil) {
				if value.IsNil() {
					continue
				}
				dbTag = cleanDbTag(dbTag, omitNil)
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
			colName = cleanDbTag(dbTag, omitEmpty, omitNil)
		} else {
			colName = strcase.ToSnake(f.Name)
		}
		cols = append(cols, table.Col(colName))
	}
	return cols
}

func cleanDbTag(tag string, tagsToClean ...string) string {
	for _, tagToClean := range tagsToClean {
		// Handle case where tag is just the partToClean
		if tag == tagToClean {
			return ""
		}

		// Handle case where tag starts with partToClean
		if strings.HasPrefix(tag, fmt.Sprintf("%s,", tagToClean)) {
			tag = strings.TrimPrefix(tag, fmt.Sprintf("%s,", tagToClean))
		}

		// Handle case where tag ends with partToClean
		if strings.HasSuffix(tag, fmt.Sprintf(",%s", tagToClean)) {
			tag = strings.TrimSuffix(tag, fmt.Sprintf(",%s", tagToClean))
		}

		// Handle case where tagToClean tag is in the middle
		if strings.Contains(tag, fmt.Sprintf(",%s,", tagToClean)) {
			tag = strings.ReplaceAll(tag, fmt.Sprintf(",%s,", tagToClean), ",")
		}
	}

	// Clean up any remaining commas
	tag = strings.Trim(tag, ",")

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

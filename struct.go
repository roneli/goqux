package goqux

import (
	"reflect"
	"strings"
	"time"

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
	skipCompare         = "skip_compare"
	// if the field is of type time.Time it will inject time.Now
	defaultNow = "now"
	// Same as default now but will inject time.Now().UTC()
	defaultNowUtc = "now_utc"
)

func encodeValues(v any, skipType string, skipZeroValues bool) map[string]SQLValuer {
	t := reflect.ValueOf(v)
	fields := reflect.VisibleFields(t.Type())
	values := make(map[string]SQLValuer)
	for _, f := range fields {
		if !f.IsExported() || strings.Contains(f.Tag.Get(tagName), skipType) {
			continue
		}
		value := t.FieldByName(f.Name)
		if !strings.Contains(f.Tag.Get(tagName), skipCompare) {
			if value.IsZero() || skipZeroValues && reflect.Zero(f.Type).Equal(value) {
				continue
			}
		}
		switch {
		case strings.Contains(f.Tag.Get(tagName), defaultNowUtc):
			values[strcase.ToSnake(f.Name)] = SQLValuer{time.Now().UTC()}
		case strings.Contains(f.Tag.Get(tagName), defaultNow):
			values[strcase.ToSnake(f.Name)] = SQLValuer{time.Now()}
		default:
			values[strcase.ToSnake(f.Name)] = SQLValuer{value.Interface()}
		}
	}
	return values
}

func getColumnsFromStruct(table exp.IdentifierExpression, s any, skipType string) []any {
	t := reflect.TypeOf(s)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	fields := reflect.VisibleFields(t)
	var cols = make([]any, 0)
	for _, f := range fields {
		if !f.IsExported() || strings.Contains(f.Tag.Get(tagName), skipType) {
			continue
		}
		if dbTag := f.Tag.Get(tagNameDb); dbTag != "" {
			cols = append(cols, table.Col(dbTag))
			continue
		} else {
			cols = append(cols, table.Col(strcase.ToSnake(f.Name)))
		}
	}
	return cols
}

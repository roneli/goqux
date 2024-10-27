package goqux

import (
	"context"
	"fmt"
	"reflect"

	"github.com/georgysavva/scany/v2/pgxscan"
)

type PaginationOptions struct {
	// PageSize per page (default: 10)
	PageSize uint
	// Use columns for key filtering, this will add a WithKeySet option to the query,
	// keys aren't validated, so make sure the names are correct or query will fail
	// if KeySet isn't set, pagination will use offset instead.
	KeySet []string
	// By Default we use ASC
	Desc bool
}

// PageIterator is a function that returns a page of results and a boolean indicating if there should be a next page or to stop iterating.
type PageIterator[T any] func(p *Paginator[T]) ([]T, bool, error)

// Paginator allows to paginate over result set of T
type Paginator[T any] struct {
	hasNext  bool
	iterator PageIterator[T]
	offset   uint
	values   []any
	stop     bool
}

func NewPaginator[T any](iterator PageIterator[T]) *Paginator[T] {
	return &Paginator[T]{
		hasNext:  true,
		iterator: iterator,
		offset:   0,
		values:   nil,
		stop:     false,
	}
}

func (p *Paginator[T]) HasMorePages() bool {
	return p.hasNext && !p.stop
}

func (p *Paginator[T]) NextPage() ([]T, error) {
	data, shouldStop, err := p.iterator(p)
	if shouldStop {
		p.stop = true
	}
	return data, err
}

func Select[T any](ctx context.Context, querier pgxscan.Querier, tableName string, options ...SelectOption) ([]T, error) {
	query, args, err := BuildSelect(tableName, new(T), options...)
	if err != nil {
		return nil, err
	}
	results := make([]T, 0)
	if err := pgxscan.Select(ctx, querier, &results, query, args...); err != nil {
		return nil, fmt.Errorf("goqux: failed to select: %w", err)
	}
	return results, nil
}

func SelectOne[T any](ctx context.Context, querier pgxscan.Querier, tableName string, options ...SelectOption) (T, error) {
	var result T
	query, args, err := BuildSelect(tableName, new(T), append(options, WithSelectLimit(1))...)
	if err != nil {
		return result, err
	}
	if err := pgxscan.Get(ctx, querier, &result, query, args...); err != nil {
		return result, fmt.Errorf("goqux: failed to select: %w", err)
	}
	return result, nil
}

func SelectPagination[T any](ctx context.Context, querier pgxscan.Querier, tableName string, paginationOptions *PaginationOptions, options ...SelectOption) (*Paginator[T], error) {
	if paginationOptions == nil {
		paginationOptions = &PaginationOptions{
			PageSize: 10,
		}
	}
	originalOptions := options
	return NewPaginator(func(p *Paginator[T]) ([]T, bool, error) {
		if paginationOptions.KeySet != nil {
			//nolint:gocritic
			cols, err := keysetColumns(paginationOptions.KeySet, new(T))
			if err != nil {
				return nil, false, err
			}
			options = append([]SelectOption{WithKeySet(cols, p.values, paginationOptions.Desc)}, originalOptions...)
			fmt.Println("keyset:OPTS", originalOptions)
		} else {
			//nolint:gocritic
			options = append([]SelectOption{WithSelectOffset(p.offset)}, originalOptions...)
		}
		results, err := Select[T](ctx, querier, tableName, append([]SelectOption{WithSelectLimit(paginationOptions.PageSize)}, options...)...)
		if err != nil {
			return nil, false, fmt.Errorf("goqux: failed to select: %w", err)
		}
		if len(results) == 0 || len(results) < int(paginationOptions.PageSize) {
			p.hasNext = false
			return results, false, nil
		}
		if len(paginationOptions.KeySet) > 0 {
			var values = make([]any, len(results))
			lastResult := results[len(results)-1]
			for i, c := range paginationOptions.KeySet {
				v := reflect.ValueOf(lastResult)
				if v.Kind() == reflect.Ptr {
					v = v.Elem()
				}
				if v.Kind() != reflect.Struct {
					return nil, false, fmt.Errorf("input is not a struct")
				}
				t := v.Type()
				for z := 0; z < t.NumField(); z++ {
					field := t.Field(z)
					fieldValue := v.Field(z)
					// Check if field name or db tag matches
					if field.Name == c || field.Tag.Get("db") == c {
						values[i] = fieldValue.Interface()
					}
				}
			}
			p.values = values
		} else {
			p.offset += paginationOptions.PageSize
		}
		return results, false, nil
	}), nil
}

func Delete[T any](ctx context.Context, querier pgxscan.Querier, tableName string, options ...DeleteOption) ([]T, error) {
	query, args, err := BuildDelete(tableName, options...)
	if err != nil {
		return nil, err
	}
	results := make([]T, 0)
	if err := pgxscan.Select(ctx, querier, &results, query, args...); err != nil {
		return nil, fmt.Errorf("goqux: failed to delete: %w", err)
	}
	return results, nil
}

func DeleteOne[T any](ctx context.Context, querier pgxscan.Querier, tableName string, options ...DeleteOption) (T, error) {
	var result T
	query, args, err := BuildDelete(tableName, append(options, WithDeleteLimit(1))...)
	if err != nil {
		return result, err
	}
	if err := pgxscan.Select(ctx, querier, &result, query, args...); err != nil {
		return result, fmt.Errorf("goqux: failed to delete: %w", err)
	}
	return result, nil
}

func Update[T any](ctx context.Context, querier pgxscan.Querier, tableName string, updateValue any, options ...UpdateOption) ([]T, error) {
	query, args, err := BuildUpdate(tableName, updateValue, options...)
	if err != nil {
		return nil, err
	}
	results := make([]T, 0)
	if err := pgxscan.Select(ctx, querier, &results, query, args...); err != nil {
		return nil, fmt.Errorf("goqux: failed to update: %w", err)
	}
	return results, nil
}

func UpdateOne[T any](ctx context.Context, querier pgxscan.Querier, tableName string, updateValue any, options ...UpdateOption) (T, error) {
	var result T
	query, args, err := BuildUpdate(tableName, updateValue, append(options, WithUpdateLimit(1))...)
	if err != nil {
		return result, err
	}
	if err := pgxscan.Select(ctx, querier, &result, query, args...); err != nil {
		return result, fmt.Errorf("goqux: failed to update: %w", err)
	}
	return result, nil
}

func Insert[T any](ctx context.Context, querier pgxscan.Querier, tableName string, insertValue any, options ...InsertOption) (*T, error) {
	var result T
	query, args, err := BuildInsert(tableName, []any{insertValue}, options...)
	if err != nil {
		return nil, err
	}
	if err := pgxscan.Get(ctx, querier, &result, query, args...); err != nil {
		if pgxscan.NotFound(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("goqux: failed to insert: %w", err)
	}
	return &result, nil
}

func InsertMany[T any](ctx context.Context, querier pgxscan.Querier, tableName string, insertValues []any, options ...InsertOption) ([]T, error) {
	query, args, err := BuildInsert(tableName, insertValues, options...)
	if err != nil {
		return nil, err
	}
	results := make([]T, 0)
	if err := pgxscan.Select(ctx, querier, &results, query, args...); err != nil {
		return nil, fmt.Errorf("goqux: failed to insert many: %w", err)
	}
	return results, nil
}

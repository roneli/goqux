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
}

type PageIterator[T any] func(p *Paginator[T]) ([]T, error)

// Paginator allows to paginate over result set of T
type Paginator[T any] struct {
	hasNext  bool
	iterator PageIterator[T]
	offset   uint
	values   []any
}

func NewPaginator[T any](iterator PageIterator[T]) *Paginator[T] {
	return &Paginator[T]{
		hasNext:  true,
		iterator: iterator,
		offset:   0,
		values:   nil,
	}
}

func (p *Paginator[T]) HasMorePages() bool {
	return p.hasNext
}

func (p *Paginator[T]) NextPage() ([]T, error) {
	return p.iterator(p)
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
	return NewPaginator(func(p *Paginator[T]) ([]T, error) {
		if paginationOptions.KeySet != nil {
			options = append(options, WithKeySet(paginationOptions.KeySet, p.values))
		} else {
			options = append(options, WithSelectOffset(p.offset))
		}
		results, err := Select[T](ctx, querier, tableName, append(options, WithSelectLimit(paginationOptions.PageSize))...)
		if err != nil {
			return nil, fmt.Errorf("goqux: failed to select: %w", err)
		}
		if len(results) == 0 || len(results) < int(paginationOptions.PageSize) {
			p.hasNext = false
		}
		if len(paginationOptions.KeySet) > 0 {
			var values = make([]any, len(results))
			lastResult := results[len(results)-1]
			reflect.ValueOf(lastResult)
			for i, c := range paginationOptions.KeySet {
				values[i] = reflect.ValueOf(lastResult).FieldByName(c).Interface()
			}
		} else {
			p.offset += paginationOptions.PageSize
		}
		return results, nil
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

func Insert[T any](ctx context.Context, querier pgxscan.Querier, tableName string, insertValue any, options ...InsertOption) (*T, error) {
	var result *T
	query, args, err := BuildInsert(tableName, []any{insertValue}, options...)
	if err != nil {
		return result, err
	}
	if err := pgxscan.Get(ctx, querier, &result, query, args...); err != nil {
		if pgxscan.NotFound(err) {
			return result, nil
		}
		return result, fmt.Errorf("goqux: failed to insert: %w", err)
	}
	return result, nil
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

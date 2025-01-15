package goqux_test

import (
	"context"
	"testing"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/jackc/pgx/v5"
	"github.com/roneli/goqux"
	"github.com/stretchr/testify/require"
)

const testPostgresURI = "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"

type User struct {
	ID       int64 `db:"id"`
	Username string
	Password string
	Email    string
}

type randomNumbers struct {
	ID     int64 `db:"id"`
	Number int64 `db:"number"`
}

func TestSelectOne(t *testing.T) {
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, testPostgresURI)
	require.Nil(t, err)
	defer func() {
		err := conn.Close(context.Background())
		require.Nil(t, err)
	}()
	tableTests := []struct {
		name           string
		options        []goqux.SelectOption
		expectedResult interface{}
	}{
		{
			name:           "simple_select",
			options:        []goqux.SelectOption{goqux.WithSelectDialect("postgres")},
			expectedResult: User{ID: 1, Username: "admin", Password: "admin", Email: "admin@acme.com"},
		},
		{
			name:           "simple_select_with_filters",
			options:        []goqux.SelectOption{goqux.WithSelectDialect("postgres"), goqux.WithSelectFilters(goqux.Column("select_users", "id").Eq(2))},
			expectedResult: User{ID: 2, Username: "user", Password: "user", Email: "user@acme.com"},
		},
	}
	for _, tt := range tableTests {
		t.Run(tt.name, func(t *testing.T) {
			model, err := goqux.SelectOne[User](ctx, conn, "select_users", tt.options...)
			require.Nil(t, err)
			require.Equal(t, tt.expectedResult, model)
		})
	}
}

func TestSelect(t *testing.T) {
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, testPostgresURI)
	require.Nil(t, err)
	defer func() {
		err := conn.Close(context.Background())
		require.Nil(t, err)
	}()
	tableTests := []struct {
		name           string
		options        []goqux.SelectOption
		expectedResult interface{}
	}{
		{
			name:           "simple_select",
			options:        []goqux.SelectOption{goqux.WithSelectDialect("postgres"), goqux.WithSelectOrder(goqu.C("id").Asc())},
			expectedResult: []User{{ID: 1, Username: "admin", Password: "admin", Email: "admin@acme.com"}, {ID: 2, Username: "user", Password: "user", Email: "user@acme.com"}},
		},
		{
			name:           "simple_select_with_filters",
			options:        []goqux.SelectOption{goqux.WithSelectDialect("postgres"), goqux.WithSelectFilters(goqux.Column("select_users", "id").Eq(2))},
			expectedResult: []User{{ID: 2, Username: "user", Password: "user", Email: "user@acme.com"}},
		},
	}
	for _, tt := range tableTests {
		t.Run(tt.name, func(t *testing.T) {
			model, err := goqux.Select[User](ctx, conn, "select_users", tt.options...)
			require.Nil(t, err)
			require.Equal(t, tt.expectedResult, model)
		})
	}
}

func TestDelete(t *testing.T) {
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, testPostgresURI)
	require.Nil(t, err)
	defer func() {
		err := conn.Close(context.Background())
		require.Nil(t, err)
	}()
	tableTests := []struct {
		name           string
		options        []goqux.DeleteOption
		expectedResult interface{}
	}{
		{
			name:           "simple_delete",
			options:        []goqux.DeleteOption{goqux.WithDeleteDialect("postgres")},
			expectedResult: []User{},
		},
		{
			name:           "delete_with_returning",
			options:        []goqux.DeleteOption{goqux.WithDeleteDialect("postgres"), goqux.WithDeleteReturningAll()},
			expectedResult: []User{},
		},
	}
	for _, tt := range tableTests {
		t.Run(tt.name, func(t *testing.T) {
			model, err := goqux.Delete[User](ctx, conn, "users", tt.options...)
			require.Nil(t, err)
			require.Equal(t, tt.expectedResult, model)
		})
	}
}

func TestInsert(t *testing.T) {
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, testPostgresURI)
	require.Nil(t, err)
	defer func() {
		err := conn.Close(context.Background())
		require.Nil(t, err)
	}()
	tableTests := []struct {
		name           string
		options        []goqux.InsertOption
		value          interface{}
		expectedResult interface{}
	}{
		{
			name:           "simple_insert",
			options:        []goqux.InsertOption{goqux.WithInsertDialect("postgres")},
			value:          User{ID: time.Now().Unix() + 5, Username: "test", Password: "test", Email: "test"},
			expectedResult: nil,
		},
		{
			name:    "insert_with_returning",
			options: []goqux.InsertOption{goqux.WithInsertDialect("postgres"), goqux.WithInsertReturning("username", "password", "email")},
			value:   User{ID: time.Now().Unix() + 4, Username: "test", Password: "test", Email: "test"},
			// ID is not return as its omitted
			expectedResult: &User{ID: 0, Username: "test", Password: "test", Email: "test"},
		},
	}
	for _, tt := range tableTests {
		t.Run(tt.name, func(t *testing.T) {
			model, err := goqux.Insert[User](ctx, conn, "users", tt.value, tt.options...)
			require.Nil(t, err)
			if model == nil {
				require.Nil(t, tt.expectedResult)
			} else {
				require.Equal(t, tt.expectedResult, model)
			}
		})
	}
}

func TestInsertMany(t *testing.T) {
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, testPostgresURI)
	require.Nil(t, err)
	defer func() {
		err := conn.Close(context.Background())
		require.Nil(t, err)
	}()
	tableTests := []struct {
		name           string
		options        []goqux.InsertOption
		value          []interface{}
		expectedResult interface{}
	}{
		{
			name:           "insert_many",
			options:        []goqux.InsertOption{goqux.WithInsertDialect("postgres")},
			value:          []any{User{ID: time.Now().Unix() + 11, Username: "test", Password: "test", Email: "test"}, User{ID: time.Now().Unix() + 13, Username: "test2", Password: "test2", Email: "test2"}},
			expectedResult: nil,
		},
		{
			name:    "insert_many_with_returning",
			options: []goqux.InsertOption{goqux.WithInsertDialect("postgres"), goqux.WithInsertReturning("username", "password", "email")},
			value:   []any{User{ID: time.Now().Unix() + 6, Username: "test", Password: "test", Email: "test"}, User{ID: time.Now().Unix() + 1, Username: "test2", Password: "test2", Email: "test2"}},
			// ID is not return as its omitted
			expectedResult: []User{{ID: 0, Username: "test", Password: "test", Email: "test"}, {ID: 0, Username: "test2", Password: "test2", Email: "test2"}},
		},
	}
	for _, tt := range tableTests {
		t.Run(tt.name, func(t *testing.T) {
			models, err := goqux.InsertMany[User](ctx, conn, "users", tt.value, tt.options...)
			require.Nil(t, err)
			if models == nil {
				require.Nil(t, tt.expectedResult)
			} else {
				require.ElementsMatch(t, tt.expectedResult, models)
			}
		})
	}
}

func TestSelectPaginationWithManyRows(t *testing.T) {
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, testPostgresURI)
	require.Nil(t, err)
	defer func() {
		err := conn.Close(context.Background())
		require.Nil(t, err)
	}()
	tableTests := []struct {
		name              string
		tableName         string
		paginationOptions *goqux.PaginationOptions
		options           []goqux.SelectOption
		expectedPages     int
	}{
		{
			name:      "paginated_select_keyset_with_many_rows",
			tableName: "random_numbers",
			paginationOptions: &goqux.PaginationOptions{
				PageSize: 10,
				KeySet:   []string{"ID"},
			},
			expectedPages: 10,
		},
	}
	for _, tt := range tableTests {
		t.Run(tt.name, func(t *testing.T) {
			paginator, err := goqux.SelectPagination[randomNumbers](ctx, conn, tt.tableName, tt.paginationOptions, tt.options...)
			require.Nil(t, err)
			totalPages := 0
			previousId := int64(0)
			for paginator.HasMorePages() {
				models, err := paginator.NextPage()
				require.Nil(t, err)
				for _, i := range models {
					require.Greater(t, i.ID, previousId)
					previousId = i.ID
				}
				totalPages += 1
			}
			require.GreaterOrEqual(t, totalPages, tt.expectedPages)
		})
	}
}

func TestSelectPagination(t *testing.T) {
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, testPostgresURI)
	require.Nil(t, err)
	defer func() {
		err := conn.Close(context.Background())
		require.Nil(t, err)
	}()
	tableTests := []struct {
		name              string
		tableName         string
		paginationOptions *goqux.PaginationOptions
		options           []goqux.SelectOption
		expectedResult    interface{}
		expectedPages     int
	}{
		{
			name:      "paginated_select_single_page",
			tableName: "select_users",
			paginationOptions: &goqux.PaginationOptions{
				PageSize: 100,
			},
			options:        []goqux.SelectOption{goqux.WithSelectDialect("postgres"), goqux.WithSelectOrder(goqu.C("id").Asc())},
			expectedResult: []User{{ID: 1, Username: "admin", Password: "admin", Email: "admin@acme.com"}, {ID: 2, Username: "user", Password: "user", Email: "user@acme.com"}},
			expectedPages:  1,
		},
		{
			name:      "paginated_select",
			tableName: "select_users",
			paginationOptions: &goqux.PaginationOptions{
				PageSize: 1,
			},
			options:        []goqux.SelectOption{goqux.WithSelectDialect("postgres"), goqux.WithSelectOrder(goqu.C("id").Asc())},
			expectedResult: []User{{ID: 1, Username: "admin", Password: "admin", Email: "admin@acme.com"}, {ID: 2, Username: "user", Password: "user", Email: "user@acme.com"}},
			expectedPages:  3,
		},
		{
			name:      "paginated_select_with_filters",
			tableName: "select_users",
			paginationOptions: &goqux.PaginationOptions{
				PageSize: 1,
			},
			options:        []goqux.SelectOption{goqux.WithSelectDialect("postgres"), goqux.WithSelectFilters(goqux.Column("select_users", "id").Eq(2)), goqux.WithSelectOrder(goqu.C("id").Asc())},
			expectedResult: []User{{ID: 2, Username: "user", Password: "user", Email: "user@acme.com"}},
			expectedPages:  2,
		},
		{
			name:      "paginated_select_keyset",
			tableName: "select_users",
			paginationOptions: &goqux.PaginationOptions{
				PageSize: 1,
				KeySet:   []string{"ID"},
			},
			options:       []goqux.SelectOption{goqux.WithSelectDialect("postgres")},
			expectedPages: 2,
		},
	}
	for _, tt := range tableTests {
		t.Run(tt.name, func(t *testing.T) {
			paginator, err := goqux.SelectPagination[User](ctx, conn, tt.tableName, tt.paginationOptions, tt.options...)
			require.Nil(t, err)
			allModels := make([]User, 0)
			totalPages := 0
			for paginator.HasMorePages() {
				models, err := paginator.NextPage()
				require.Nil(t, err)
				allModels = append(allModels, models...)
				totalPages += 1
			}
			if tt.expectedResult != nil {
				require.Equal(t, tt.expectedResult, allModels)
			}
			require.GreaterOrEqual(t, totalPages, tt.expectedPages)
		})
	}
}

func TestPaginateQueryByKeySetWithEmptyKeySet(t *testing.T) {
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, testPostgresURI)
	require.Nil(t, err)
	defer func() {
		err := conn.Close(context.Background())
		require.Nil(t, err)
	}()
	_, err = goqux.QueryKeySetPagination[User](ctx, conn, goqu.From("users"), 10, []string{})
	require.Error(t, err)
}

func TestPaginateQueryByKeySetWithNoResults(t *testing.T) {
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, testPostgresURI)
	require.Nil(t, err)
	defer func() {
		err := conn.Close(context.Background())
		require.Nil(t, err)
	}()
	paginator, err := goqux.QueryKeySetPagination[User](ctx, conn, goqu.Dialect("postgres").From("users").Where(goqu.C("id").Eq(999)), 10, []string{"ID"})
	require.Nil(t, err)
	require.NotNil(t, paginator)
	results, err := paginator.NextPage()
	require.Nil(t, err)
	require.Empty(t, results)
}

func TestPaginateQueryByKeySetWithMultiplePages(t *testing.T) {
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, testPostgresURI)
	require.NoError(t, err)
	defer func() {
		err := conn.Close(context.Background())
		require.NoError(t, err)
	}()
	_, err = goqux.Delete[User](ctx, conn, "users")
	require.NoError(t, err)
	random := time.Now().Unix()
	_, err = goqux.InsertMany[User](ctx, conn, "users", []any{
		User{ID: random + 200, Username: "test2", Password: "test2", Email: "test2"},
		User{ID: random + 100, Username: "test", Password: "test", Email: "test"},
		User{ID: random + 300, Username: "test3", Password: "test3", Email: "test3"},
	})
	expectedUserNames := []string{"test", "test2", "test3"}
	require.NoError(t, err)
	paginator, err := goqux.QueryKeySetPagination[User](ctx, conn, goqu.Dialect("postgres").Select("id", "username").From("users"), 1, []string{"ID"})
	require.NoError(t, err)
	require.NotNil(t, paginator)
	page := 0
	for paginator.HasMorePages() {
		u, err := paginator.NextPage()
		require.NoError(t, err)
		if len(u) == 0 {
			break
		}
		require.Equal(t, len(u), 1)
		require.Equal(t, u[0].Username, expectedUserNames[page])
		page++
	}
	require.Equal(t, page, 3)
}

func TestStopOnPagination(t *testing.T) {
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, testPostgresURI)
	require.Nil(t, err)
	defer func() {
		err := conn.Close(context.Background())
		require.Nil(t, err)
	}()
	paginatorMock := goqux.NewPaginator(func(p *goqux.Paginator[User]) ([]User, bool, error) {
		return []User{}, true, nil
	})
	countPages := 0
	for paginatorMock.HasMorePages() {
		_, err := paginatorMock.NextPage()
		require.Nil(t, err)
		countPages++
	}
	require.Equal(t, 1, countPages)
}

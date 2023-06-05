```
                           ____  ___
   ____   ____   ________ _\   \/  /
  / ___\ /  _ \ / ____/  |  \     / 
 / /_/  >  <_> < <_|  |  |  /     \ 
 \___  / \____/ \__   |____/___/\  \
/_____/            |__|          \_/
```

GoquX is a lightweight wrapper library for [goqu](https://github.com/doug-martin/goqu), 
designed to simplify the process of building CRUD queries, implementing pagination, and struct scanning [scany](https://github.com/georgysavva/scany).

## Features

- Builder helpers for **select**/**insert**/**update**/**delete** queries, auto adding columns and serialization of rows, tags for skipping columns/setting default values.
- Query Execution support, with **Pagination** for offset/limit and keyset pagination.
- **Automatic** scanning into structs using [scany](https://github.com/georgysavva/scany). 
- **Customizable** builder options, allowing you to easily extend the builder options.

## Why?

There is much debate in Golang about the best way to handle database queries. Some prefer ORM libraries like GORM,
while others prefer to use query builders like goqu, and of course, there are those who prefer to write raw SQL queries.

Personally, I usually like to use query builders as they offer a good balance, and use raw queries when it's very complex query.

I wrote GoquX because I found myself writing the same code over and over again for simple queries,  and I wanted to simplify 
the process of building CRUD queries, implementing pagination, and struct scanning.

GoquX is not a replacement for goqu, but rather a lightweight wrapper that simplifies the process of using it.


## Installation

To use GoquX in your Go project, you need to have Go installed and set up on your machine. Then, run the following command to add GoquX as a dependency:

```bash
go get github.com/roneli/goqux
```

## Examples

### Pagination

`goqux` adds a convenient pagination function allowing us to scan the results into a slice of structs, add filters, ordering, 
and extend the query with any other goqu function.

Pagination currently supports offset/limit or keyset pagination.

```go 
conn, err := pgx.Connect(ctx, "postgres://postgres:postgres@localhost:5432/postgres")
if err != nil {
    log.Fatal(err)
}
paginator, err := goqux.SelectPagination[User](ctx, conn, "users", &goqux.PaginationOptions{ PageSize: 100}, goqux.WithSelectFilters(goqux.Column("users", "id").Eq(2)))
for paginator.HasMorePages() {
    users, err := paginator.NextPage()
    ...
}
```
### Select/Insert/Update/Delete Query building

`goqux` adds select/insert/update/delete simple utilities to build queries.

#### Select Builder

```go
type User struct {
    ID        int64     `db:"id"`
    Name      string    `db:"name"`
    Email     string    `db:"email"`
    CreatedAt time.Time `db:"created_at"`
    UpdatedAt time.Time `db:"updated_at"`
    FieldToSkip string  `goqux:"skip_select"`
}
// Easily extend the query with any other goqux function optional functions that get access to the query builder.
// use goqux:"skip_select" to skip a field in the select query.
sql, args, err := goqux.BuildSelect("table_to_select", User{},
    goqux.WithSelectFilters(goqux.Column("table_to_select", "id").Gt(2)),
    goqux.WithSelectOrder(goqux.Column("table_to_select", "id").Desc()),
    goqux.WithSelectLimit(10),
)
```

#### Insert Builder

```go
type User struct {
    ID        int64     `db:"id"`
    Name      string    `db:"name"`
    Email     string    `db:"email"`
    CreatedAt time.Time `goqux:"now,skip_update"`
    UpdatedAt time.Time `goqux:"now_utc"`
    FieldToSkip string  `goqux:"skip_insert"`
}
// use goqux:"now" to set the current time in the insert query for CreatedAt, and goqux:"now_utc" to set the current time in UTC for UpdatedAt. 
sql, args, err := goqux.BuildInsert("table_to_insert", User{ID: 5, Name: "test", Email: "test@test.com"}, goqu.WithReturningAll()),
)
```

### Select/Insert/Update/Delete Executions

`goqux` adds select/insert/update/delete functions to execute simple queries.

#### SelectOne
```go
    user, err := goqux.SelectOne[User](ctx, conn, "users", goqux.WithSelectFilters(goqux.Column("users", "id").Eq(2)))
```

#### Select
```go
    user, err := goqux.Select[User](ctx, conn, "users",  goqux.WithSelectOrder(goqu.C("id").Asc()))
```

#### Insert
```go

```

### Easily extend builder options
You can define any custom option you want to extend the builder options, for example, if you want to add a group by option you can do the following:
```go
func WithSelectGroupBy(columns ...any) SelectOption {
	return func(_ exp.IdentifierExpression, s *goqu.SelectDataset) *goqu.SelectDataset {
		return s.GroupBy(columns...)
	}
}
```

You can add these options to any of the insert/update/delete/select functions.


##### For more examples check the tests.
// 26.04.15 11:40
// (c) Dmitriy Blokhin (sv.dblokhin@gmail.com), www.webjinn.ru

package dbwrapper

import (
	"database/sql"
	"errors"
	_ "github.com/go-sql-driver/mysql"
	"strings"
	"context"
)

// Database db-instance
type Database struct {
	driver *sql.DB
	prefixer *strings.Replacer
	escaper *strings.Replacer
}

var (
	// errors
	errNoDatabase = errors.New("dbwrapper: no database instance")
	errColumns = errors.New("sqlfetch: get columns error")
)

// Query executes sql query wih arguments
func (db Database) Query(sql string, args ...interface{}) []map[string]string {
	if db.prefixer != nil {
		sql = db.prefixer.Replace(sql)
	}

	if db.driver == nil {
		panic(errNoDatabase)
	}

	rows, err := db.driver.Query(sql, args...)
	if err != nil {
		panic(err)
	}

	defer rows.Close()
	return db.sqlFetch(rows)
}

// Row executes SQL query and returns a row
func (db Database) Row(sql string, args ...interface{}) map[string]string {
	// TODO: dont use db.Query
	res := db.Query(sql, args...)
	if len(res) > 0 {
		return res[0]
	}

	return map[string]string{}
}

// Result executes SQL query and returns a field
func (db Database) Result(sql string, args ...interface{}) string {
	res := db.Row(sql, args...)
	for _, val := range res {
		return val
	}

	return ""
}

// Exec executes no data SQL query
func (db Database) Exec(sql string, args ...interface{}) sql.Result {
	if db.prefixer != nil {
		sql = db.prefixer.Replace(sql)
	}

	if db.driver == nil {
		panic(errNoDatabase)
	}

	res, err := db.driver.Exec(sql, args...)
	if err != nil {
		panic(err)
	}

	return res
}

// ExecId executes no data SQL query and returns inserted id
func (db Database) ExecId(sql string, args ...interface{}) int64 {
	if db.prefixer != nil {
		sql = db.prefixer.Replace(sql)
	}

	if db.driver == nil {
		panic(errNoDatabase)
	}

	res, err := db.driver.Exec(sql, args...)

	if err != nil {
		panic(err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		panic(err)
	}

	return id
}

// EscapeString escapes string
func (db Database) EscapeString(s string) string {
	return db.escaper.Replace(s)
}

func (db Database) sqlFetch(Rows *sql.Rows) []map[string]string {

	columns, err := Rows.Columns()
	if err != nil {
		panic(err)
	}

	if len(columns) == 0 {
		panic(err)
	}

	values := make([]sql.RawBytes, len(columns))
	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	resultQuery := make([]map[string]string, 0)
	for Rows.Next() {
		newRow := make(map[string]string)

		if err := Rows.Scan(scanArgs...); err != nil {
			panic(err)
		}
		var value string
		for i, col := range values {
			if col == nil {
				value = "NULL"
			} else {
				value = string(col)
			}
			newRow[columns[i]] = value
		}
		resultQuery = append(resultQuery, newRow)
	}

	return resultQuery
}

type key int

const keyDB key = iota

// New creates db wrapper
func New(driver, source, prefix string) *Database {
	var err error

	db := new(Database)
	db.prefixer = strings.NewReplacer("#__", prefix)
	db.escaper = strings.NewReplacer(`'`, `\'`, `\`, `\\`, `"`, `\"`)

	if db.driver, err = sql.Open(driver, source); err != nil {
		panic(err)
	}

	if err = db.driver.Ping(); err != nil {
		panic(err)
	}

	return db
}

// NewFromDB returns dbwrapper from active *sql.Database
func NewFromDB(drv *sql.DB, prefix string) *Database {
	return &Database{
		driver: drv,
		prefixer: strings.NewReplacer("#__", prefix),
		escaper: strings.NewReplacer(`'`, `\'`, `\`, `\\`, `"`, `\"`),
	}
}

// NewContext create new context with db instance
func NewContext(ctx context.Context, driver, source, prefix string) context.Context {
	return context.WithValue(ctx, keyDB, New(driver, source, prefix))
}

// DB return instance from context
func DB(ctx context.Context) *Database {
	value, ok := ctx.Value(keyDB).(*Database)
	if !ok {
		panic(errNoDatabase)
	}

	return value
}
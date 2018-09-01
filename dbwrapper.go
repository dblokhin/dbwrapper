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
	errNoDatabase = errors.New("no database instance")
	errColumns = errors.New("sqlfetch: get columns error")
)

// Query executes sql query wih arguments
func (db Database) Query(sql string, args ...interface{}) ([]map[string]string, error) {
	if db.prefixer != nil {
		sql = db.prefixer.Replace(sql)
	}

	if db.driver == nil {
		return []map[string]string{}, errNoDatabase
	}

	rows, err := db.driver.Query(sql, args...)
	if err != nil {
		return []map[string]string{}, err
	}

	defer rows.Close()
	return db.sqlFetch(rows)
}

// Row executes sql query and returns a row
func (db Database) Row(sql string, args ...interface{}) (map[string]string, error) {
	// TODO: dont use db.Query
	res, err := db.Query(sql, args...)
	if err != nil {
		return map[string]string{}, err
	}

	if len(res) > 0 {
		return res[0], nil
	}

	return map[string]string{}, nil
}

// Result executes sql query and returns a field
func (db Database) Result(sql string, args ...interface{}) (string, error) {
	res, err := db.Row(sql, args...)
	if err != nil {
		return "", err
	}

	for _, val := range res {
		return val, nil
	}

	return "", nil
}

// Exec executes nodata sql query
func (db Database) Exec(sql string, args ...interface{}) (sql.Result, error) {
	if db.prefixer != nil {
		sql = db.prefixer.Replace(sql)
	}

	return db.driver.Exec(sql, args...)
}

// ExecId executes nodata sql query and returns inserted id
func (db Database) ExecId(sql string, args ...interface{}) (int64, error) {
	if db.prefixer != nil {
		sql = db.prefixer.Replace(sql)
	}

	if res, err := db.driver.Exec(sql, args...); err != nil {
		return 0, err
	} else {
		return res.LastInsertId()
	}
}

// EscapeString escapes string
func (db Database) EscapeString(s string) string {
	return db.escaper.Replace(s)
}

func (db Database) sqlFetch(Rows *sql.Rows) ([]map[string]string, error) {

	columns, err := Rows.Columns()
	if err != nil {
		return []map[string]string{}, err
	}

	if len(columns) == 0 {
		return []map[string]string{}, errColumns
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
			return []map[string]string{}, err
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

	return resultQuery, nil
}

type key int

const keyDB key = iota

// New creates db wrapper
func New(driver, source, prefix string) (db *Database, err error) {
	db = new(Database)
	db.prefixer = strings.NewReplacer("#__", prefix)
	db.escaper = strings.NewReplacer(`'`, `\'`, `\`, `\\`, `"`, `\"`)

	if db.driver, err = sql.Open(driver, source); err != nil {
		return
	}

	if err = db.driver.Ping(); err != nil {
		return
	}

	return
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
func NewContext(ctx context.Context, driver, source, prefix string) (context.Context, error) {
	if db, err := New(driver, source, prefix); err != nil {
		return ctx, err
	} else {
		return context.WithValue(ctx, keyDB, db), nil
	}
}

// DB return instance from context
func DB(ctx context.Context) (*Database, bool) {
	value, ok := ctx.Value(keyDB).(*Database)
	return value, ok
}
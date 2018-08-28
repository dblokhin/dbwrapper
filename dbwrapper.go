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
}

var (
	prefixer *strings.Replacer
	escaper *strings.Replacer
)

func (db Database) Query(sql string, args ...interface{}) ([]map[string]string, error) {
	// Prefix
	sql = prefixer.Replace(sql)

	if rows, err := db.driver.Query(sql, args...); err != nil {
		return []map[string]string{}, err
	} else {
		defer rows.Close()
		return sqlFetch(rows)
	}
}

func (db Database) Row(sql string, args ...interface{}) (map[string]string, error) {
	res, err := db.Query(sql, args...)
	if err != nil {
		return map[string]string{}, err
	}

	if len(res) > 0 {
		return res[0], nil
	}

	return map[string]string{}, nil
}

func (db Database) Result(sql string, args ...interface{}) (string, error) {
	res, err := db.Query(sql, args...)
	if err != nil {
		return "", err
	}

	if len(res) > 0 {
		for _, val := range res[0] {
			return val, nil
		}
	}

	return "", nil
}

func (db Database) Exec(sql string, args ...interface{}) (sql.Result, error) {
	// Prefix
	sql = prefixer.Replace(sql)

	return db.driver.Exec(sql, args...)
}

func (db Database) ExecId(sql string, args ...interface{}) (int64, error) {
	// Prefix
	sql = prefixer.Replace(sql)

	if res, err := db.driver.Exec(sql, args...); err != nil {
		return 0, err
	} else {
		return res.LastInsertId()
	}
}

func (db Database) EscapeString(s string) string {
	return escaper.Replace(s)
}

func sqlFetch(Rows *sql.Rows) ([]map[string]string, error) {

	columns, err := Rows.Columns()
	if err != nil {
		return []map[string]string{}, err
	}

	if len(columns) == 0 {
		return []map[string]string{}, errors.New("sqlfetch: get columns error")
	}

	values := make([]sql.RawBytes, len(columns))
	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	data := make([]map[string]string, 0)

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
		data = append(data, newRow)
	}

	return data, nil
}

type key int

const keyDB key = iota

// New creates dbwrapper
func New(driver, source, prefix string) (db Database, err error) {

	db.driver, err = sql.Open(driver, source)
	if err != nil {
		return
	}

	err = db.driver.Ping()
	if err != nil {
		return
	}

	prefixer = strings.NewReplacer("#__", prefix)
	escaper = strings.NewReplacer(`'`, `\'`, `\`, `\\`, `"`, `\"`)

	return db, err
}

// NewFromDB returns dbwrapper from active *sql.Database
func NewFromDB(drv *sql.DB, prefix string) Database {
	db := Database{
		driver: drv,
	}

	prefixer = strings.NewReplacer("#__", prefix)
	escaper = strings.NewReplacer(`'`, `\'`, `\`, `\\`, `"`, `\"`)

	return db
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
func DB(ctx context.Context) (Database, bool) {
	value, ok := ctx.Value(keyDB).(Database)
	return value, ok
}
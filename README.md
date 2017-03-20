# Simple databases Go wrapper
Simple databases Golang wrapper that all requests are returned as `[]map[string]string`. Very easy but isnt cool efficient. Easy to use with golang `"html/template"` package.

## How to use
```go
import (
	"github.com/dblokhin/config"
	database "github.com/dblokhin/dbwrapper"
	// and others packages ...
)

func main()  {
	// load config
	app := config.NewContext(context.Background(), "config.json")
	config, _ := config.FromContext(app)    // if no config <- will be panic
	dbsource := config.GetString("db.source")
	dbprefix := conf.GetString("db.prefix")

	if config.GetBool("db.enabled") {
		// connect to db
		app, _ = database.NewContext(app, "mysql", dbsource, dbprefix)
		
		if db, ok := database.FromContext(app); !ok {
			panic("no database")
		} else {
			// Settings up
			db.Exec("SET NAMES utf8")
			db.Exec("SET SESSION sql_mode = 'TRADITIONAL'")
		}
	}
	// ...

	param1 := "NOT safe sql string"
	param1 := 15

	result, err := db.Query("SELECT a FROM #__b WHERE c = ?, d = ?", param1, param2)
}
```

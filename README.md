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
	config := config.Config(app)    // if no config <- will be panic
	dbsource := config.GetString("db.source")
	dbprefix := conf.GetString("db.prefix")

	if config.GetBool("db.enabled") {
		// connect to db
		app = database.NewContext(app, "mysql", dbsource, dbprefix)
		
		db := database.FromContext(app)
        // Settings up
        db.Exec("SET NAMES utf8")
        db.Exec("SET SESSION sql_mode = 'TRADITIONAL'")
	}
	
	// ... another place
	param1 := "NOT safe sql string"
	param1 := 15

	result, err := db.Query("SELECT a FROM #__b WHERE c = ?, d = ?", param1, param2)
}
```

## Why does it panic?
**The package doesn't return any errors, but it does panic.** In my opinion it's good *error handling* way that allows easy coding & good concentrating on that.

Lets me describe this point. If you use function that returns an errors, you have to (must) check every time annoying `if err != nil {}`. A good way of easy coding, in my opinion, is that functions could return only the useful values or only errors, like this:

`func SomeFunc() int | error`

and handles values and errors separately. Directly Golang doesn't allow it, but `defer` & `panic` allow us it. Just few examples:

#### Before. Consider some `initial` function:
```go
	// create app instance & load config
	app, err := webapp.New()
	if err != nil {
		return nil, err
	}

	app, err = config.NewContext(app, "config.json")
	if err != nil {
		return nil, err
	}

	conf, err := config.Config(app)
	if err != nil {
		return nil, err
	}
    // ...
```
The some caller is: 
```go
	// initiate the app
	app, err := someInitial()
	if err != nil {
		log.Println(err)
		os.Exit(someErrCode)
	}
```
In most cases `error` just means return function. And caller can checks the `err` again & return it too...
#### After. New nice code:
```go
	app := webapp.New()
	app = config.NewContext(app, "config.json")
	conf := config.Config(app)
```
#### But How to handle Errors?
Error handler in caller (or may be caller of caller):
```go
	// error handler
	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
            os.Exit(someErrCode)
		}
	}()
    
    // initiate the app
    app := someInitial()
```
```go
	// error handler
	defer func() {
		if ret := recover(); ret != nil {
			cli.Status(http.StatusBadRequest)

			switch err := ret.(type) {
			case error:
				json.NewEncoder(cli).Encode(jsonError{
					Msg: err.Error(),
				})
			case string:
				json.NewEncoder(cli).Encode(jsonError{
					Msg: err,
				})

			default:
				log.Println("unkown error panic")
			}
		}
	}()
```

We can manipulate error values in `recover()`, we can place error handlers package.

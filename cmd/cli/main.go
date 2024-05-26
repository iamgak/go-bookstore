package main

import (
	"flag"
	// "fmt"
	"database/sql"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql" // sql pool register

	"test.iamgak.net/models"
)

type application struct {
	infoLog  *log.Logger
	errorLog *log.Logger
	db       *sql.DB
	books    *models.BookModel
	user     *models.UserModel
	review   *models.ReviewModel
}

func main() {

	addr := flag.String("addr", ":4000", "HTTP network address")
	dsn := flag.String("dsn", "root:@/bookreview?parseTime=true", "MySQL data source name")
	flag.Parse()
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)
	db, err := openDB(*dsn)
	if err != nil {
		errorLog.Fatal(err)
	}
	defer db.Close()

	// And add it to the application dependencies.
	app := &application{
		errorLog: errorLog,
		infoLog:  infoLog,
		db:       db,
		user:     &models.UserModel{DB: db},
		books:    &models.BookModel{DB: db},
		review:   &models.ReviewModel{DB: db},
		// validator: &validator.Validator{Errors: make(map[string]string)},
	}

	srv := &http.Server{
		Addr:     *addr,
		ErrorLog: errorLog,
		Handler:  app.routes(),
	}

	infoLog.Printf("Starting server on %s", *addr)
	err = srv.ListenAndServe()
	errorLog.Fatal(err)
}

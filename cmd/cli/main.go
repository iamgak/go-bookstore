package main

import (
	"crypto/tls"
	"database/sql"
	"flag"
	_ "github.com/go-sql-driver/mysql" // sql pool register
	"log"
	"net/http"
	"os"
	"test.iamgak.net/models"
	"time"
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
	dsn := flag.String("dsn", "root:@/bookstore?parseTime=true", "MySQL data source name")
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
	}

	tlsConfig := &tls.Config{
		CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
	}
	srv := &http.Server{
		Addr:      *addr,
		ErrorLog:  errorLog,
		TLSConfig: tlsConfig,
		// MaxHeaderBytes: 524288, // 0.5MB Max header size per request
		IdleTimeout:  time.Minute, // conncection close after 1 minute it do again handshake or something
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		Handler:      app.routes(),
	}

	infoLog.Printf("Starting server on %s", *addr)
	err = srv.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem")
	errorLog.Fatal(err)
}

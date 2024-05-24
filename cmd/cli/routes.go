package main

import "net/http"

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", app.Home)
	mux.HandleFunc("/book/listing", app.BooksListing)
	mux.HandleFunc("/book/info", app.BooksInfo)
	// mux.HandleFunc("/book/request", app.BooksInfo)
	mux.HandleFunc("/book/create", app.AddBooks)
	mux.HandleFunc("/user/register", app.UserRegister)
	mux.HandleFunc("/user/activation", app.UserActivation)
	mux.HandleFunc("/user/login", app.UserLogin)
	mux.HandleFunc("/user/logout", app.UserLogout)
	return secureHeaders(mux)
}

package main

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()
	router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.notFound(w)
	})

	router.HandlerFunc(http.MethodGet, "/", app.Home)
	router.HandlerFunc(http.MethodGet, "/book/listing", app.BooksListing)
	router.HandlerFunc(http.MethodGet, "/book/info/:isbn", app.BooksInfo)
	router.HandlerFunc(http.MethodPost, "/book/create", app.AddBooks)
	router.HandlerFunc(http.MethodPost, "/user/register", app.UserRegister)
	router.HandlerFunc(http.MethodPost, "/user/forget_password/", app.ForgetPasswordPost)
	router.HandlerFunc(http.MethodPost, "/user/new_password/:uri", app.NewPasswordPost)
	router.HandlerFunc(http.MethodGet, "/user/activation/:uri", app.UserActivation)
	router.HandlerFunc(http.MethodPost, "/user/login", app.UserLogin)
	router.HandlerFunc(http.MethodPost, "/user/logout", app.UserLogout)
	return router
}

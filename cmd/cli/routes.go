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

	//home related routes
	router.HandlerFunc(http.MethodGet, "/", app.Home)
	//books related routes
	router.HandlerFunc(http.MethodGet, "/book/listing", app.BookListing)    // all the book listing
	router.HandlerFunc(http.MethodGet, "/book/search/:isbn/", app.BookInfo) // review of given isbn
	router.HandlerFunc(http.MethodPost, "/book/create", app.AddBook)        // create a new book info different isbn
	//review related routes
	router.HandlerFunc(http.MethodGet, "/review/listing", app.ReviewListing)      // all the reviews
	router.HandlerFunc(http.MethodGet, "/myreview/", app.MyReview)                // review of logged in user
	router.HandlerFunc(http.MethodGet, "/review/search/:isbn/", app.ReviewSearch) // review of given isbn
	router.HandlerFunc(http.MethodPost, "/review/create", app.AddReview)          // create review
	router.HandlerFunc(http.MethodGet, "/review/delete/:id", app.DeleteReview)    // delete your own review
	//user related routes
	router.HandlerFunc(http.MethodPost, "/user/forget_password/", app.ForgetPasswordPost) // to create forget password request
	router.HandlerFunc(http.MethodPost, "/user/register", app.UserRegister)               // to register
	router.HandlerFunc(http.MethodPost, "/user/new_password/:uri", app.NewPasswordPost)   //after forget password req uri created to change passw
	router.HandlerFunc(http.MethodGet, "/user/activation/:uri", app.UserActivation)       // after registration uri created authentication
	router.HandlerFunc(http.MethodPost, "/user/login", app.UserLogin)                     // login
	router.HandlerFunc(http.MethodPost, "/user/logout", app.UserLogout)                   // logout
	return app.logRequest(secureHeaders(router))
}

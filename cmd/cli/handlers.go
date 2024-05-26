package main

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"net/http"

	// "regexp"
	"github.com/julienschmidt/httprouter"
	"strconv"
	"test.iamgak.net/models"
	"test.iamgak.net/validator"

	// "strings"
	"time"
)

type Message struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
}

func (app *application) Home(w http.ResponseWriter, r *http.Request) {
	resp := Message{
		Status:  true,
		Message: "Welcome to our Bookstore, Website",
	}

	app.sendJSONResponse(w, 200, &resp)
}

// bookListing related handlers
func (app *application) BookListing(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	isbn := params.ByName("isbn")
	bks, err := app.books.GetBookByIsbn(isbn)
	if err != nil {
		app.errorLog.Print("internal server error")
		app.CustomError(w, "Internal Server Error", 500)
		return
	}

	app.sendJSONResponse(w, 200, bks)
}

// listing
func (app *application) ReviewListing(w http.ResponseWriter, r *http.Request) {
	bks, err := app.review.ReviewListing()
	if err != nil {
		app.errorLog.Print("internal server error")
		app.CustomError(w, "Internal Server Error", 500)
		return
	}

	app.sendJSONResponse(w, 200, bks)
}

func (app *application) MyReview(w http.ResponseWriter, r *http.Request) {
	uid := app.ValidToken(w, r)
	if uid == 0 {
		app.notFound(w)
		return
	}

	bks, err := app.review.MyReview(uid)
	if err != nil {
		app.errorLog.Print("internal server error")
		app.CustomError(w, "Internal Server Error", 500)
		return
	}

	app.sendJSONResponse(w, 200, bks)
}

func (app *application) ReviewSearch(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	isbn := params.ByName("isbn")
	bks, err := app.review.GetReviewByIsbn(isbn)
	if err != nil {
		app.errorLog.Print("internal server error")
		app.CustomError(w, "Internal Server Error", 500)
		return
	}

	app.sendJSONResponse(w, 200, bks)
}

func (app *application) DeleteReview(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	review_id, err := strconv.ParseInt(params.ByName("id"), 10, 32)
	if err != nil {
		app.notFound(w)
		return
	}

	uid := app.ValidToken(w, r)
	if uid == 0 {
		app.CustomError(w, "Authorization failed. Please provide a valid bearer token to access this resource", 401)
		return
	}

	err = app.review.DeleteReview(int(review_id), int(uid))
	if err != nil {
		app.notFound(w)
		return
	}

	resp := app.sendMessage(true, "Review Deleted")
	app.sendJSONResponse(w, 200, resp)
}

func (app *application) BookInfo(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	isbn := params.ByName("isbn")
	validator := &validator.Validator{
		Errors: make(map[string]string),
	}
	validator.CheckField(validator.NotBlank(isbn), "ISBN", "Empty, ISBN field")

	if !validator.Valid() {
		app.sendJSONResponse(w, 200, validator.Errors)
		return
	}

	info, err := app.books.GET(isbn)
	if err != nil {
		app.CustomError(w, "Internal Server Error", 500)
		return
	}

	app.sendJSONResponse(w, 200, info)
}

func (app *application) AddReview(w http.ResponseWriter, r *http.Request) {
	uid := app.ValidToken(w, r)
	if uid == 0 {
		app.CustomError(w, "Authorization failed. Please provide a valid bearer token to access this resource", 401)
		return
	}

	validator := &validator.Validator{
		Errors: make(map[string]string),
	}
	isbn := r.FormValue("isbn")
	title := r.FormValue("title")
	// rating := r.FormValue("rating")
	descriptions := r.FormValue("descriptions")
	price, err := strconv.ParseFloat(r.FormValue("price"), 32)
	if err != nil {
		validator.Errors["price"] = "Invalid price value"
	}
	rating, err := strconv.ParseFloat(r.FormValue("rating"), 32)
	if err != nil {
		validator.Errors["rating"] = "Invalid rating value"
	}

	validator.CheckField(validator.NotBlank(isbn), "isbn", "Please, fill the isbn field")
	validator.CheckField(validator.NotBlank(descriptions), "descriptions", "Please, fill the descriptions field")
	validator.CheckField(validator.NotBlank(title), "title", "Please, fill the title field")

	if validator.Errors["isbn"] == "" {
		validator.CheckField(validator.MaxChars(isbn, 20), "isbn", "Please, fill the ISBN shorter than 20")
	}

	if validator.Errors["descriptions"] == "" {
		validator.CheckField(validator.MaxChars(descriptions, 100), "isbn", "Please, fill the ISBN shorter than 100")
	}

	if validator.Errors["title"] == "" {
		validator.CheckField(validator.MaxChars(title, 50), "title", "Please, fill the TITLE shorter than 50")
	}

	if !validator.Valid() {
		app.sendJSONResponse(w, 200, validator)
		return
	}

	review := &models.Review{
		Isbn:         isbn,
		Title:        title,
		Rating:       float32(rating),
		Price:        float32(price),
		Descriptions: descriptions,
		Uid:          uid,
	}
	err = app.review.CreateReview(review)
	if err != nil {
		app.CustomError(w, "Server Issue", 500)
		return
	}

	resp := app.sendMessage(true, "Review Saved")
	app.sendJSONResponse(w, 200, resp)
}

func (app *application) AddBook(w http.ResponseWriter, r *http.Request) {
	uid := app.ValidToken(w, r)
	if uid == 0 {
		app.CustomError(w, "Authorization failed. Please provide a valid bearer token to access this resource", 401)
		return
	}

	isbn := r.FormValue("isbn")
	title := r.FormValue("title")
	author := r.FormValue("author")
	genre := r.FormValue("genre")
	descriptions := r.FormValue("descriptions")
	validator := &validator.Validator{
		Errors: make(map[string]string),
	}

	validator.CheckField(validator.NotBlank(isbn), "isbn", "Please, fill the isbn field")
	validator.CheckField(validator.NotBlank(genre), "genre", "Please, fill the genre field")
	validator.CheckField(validator.NotBlank(descriptions), "descriptions", "Please, fill the descriptions field")
	validator.CheckField(validator.NotBlank(author), "author", "Please, fill the author field")
	validator.CheckField(validator.NotBlank(title), "title", "Please, fill the title field")

	if validator.Errors["isbn"] == "" {
		validator.CheckField(validator.MaxChars(isbn, 20), "isbn", "Please, fill the ISBN shorter than 20")
	}

	if validator.Errors["descriptions"] == "" {
		validator.CheckField(validator.MaxChars(descriptions, 100), "isbn", "Please, fill the ISBN shorter than 100")
	}

	if validator.Errors["author"] == "" {
		validator.CheckField(validator.MaxChars(author, 50), "author", "Please, fill the AUTHOR shorter than 50")
	}
	if validator.Errors["title"] == "" {
		validator.CheckField(validator.MaxChars(title, 50), "title", "Please, fill the TITLE shorter than 50")
	}

	price, err := strconv.ParseFloat(r.FormValue("price"), 32)
	if err != nil {
		validator.Errors["price"] = "Invalid price value"
	}

	if validator.Valid() {
		book_id := app.books.BookExist(isbn)
		if book_id {
			validator.Errors["isbn"] = "isbn already Exist"
		}
	}

	if !validator.Valid() {
		app.sendJSONResponse(w, 200, validator)
		return
	}

	bookInfo := &models.Book{
		ISBN:         isbn,
		Title:        title,
		Genre:        genre,
		Price:        float32(price),
		Descriptions: descriptions,
		Author:       author,
	}

	err = app.books.CreateBook(bookInfo)
	if err != nil {
		app.CustomError(w, "Server Issue1", 500)
		return
	}

	resp := app.sendMessage(true, "Book Record Saved, Sucessfully")
	app.sendJSONResponse(w, 200, resp)
}

func (app *application) UserRegister(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.errorLog.Println("Error parsing form data:", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")
	repeatPassword := r.FormValue("repeatPassword")
	validator := &validator.Validator{
		Errors: make(map[string]string),
	}

	validator.CheckField(validator.NotBlank(email), "email", "Please, fill the email field")
	validator.CheckField(validator.NotBlank(password), "password", "Please, fill the password field")
	validator.CheckField(validator.NotBlank(repeatPassword), "repeatPassword", "Please, fill the repeat password field")
	if validator.Errors["password"] == "" {
		validator.CheckField(validator.ValidPassword(password), "password", "Password should be greater than 6 character must contain alphanumeric char, one special char")
	}

	if validator.Errors["email"] == "" {
		validator.CheckField(validator.ValidEmail(email), "email", "Invalid Email Format")
	}
	if app.user.EmailExist(email) != 0 {
		validator.CheckField(false, "email", "Email already registered")
	}

	if validator.Errors["repeatPassword"] == "" && validator.Errors["password"] == "" {
		if password != repeatPassword {
			validator.Errors["repeatPassword"] = "Password not matched"
		}
	}

	if !validator.Valid() {
		app.sendJSONResponse(w, 200, validator)
		return
	}

	uri := app.generateHash(r.RemoteAddr, r.URL.Port())
	exist, err := app.user.InsertUser(email, password, uri)
	if err != nil || !exist {
		app.serverError(w, err)
		return
	}

	resp := Message{
		Status:  true,
		Message: "Registration Successfull ",
	}

	app.sendJSONResponse(w, 200, resp)
}

func (app *application) UserActivation(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	uri := params.ByName("uri")
	if !app.user.ValidURI(uri) {
		app.notFound(w)
		return
	}

	app.infoLog.Print("hello")
	err := app.user.AccountActivate(uri)
	if err != nil {
		app.notFound(w)
		return
	}

	resp := app.sendMessage(true, "Your account has been Verified.")
	app.sendJSONResponse(w, 200, resp)
}

func (app *application) UserLogin(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.errorLog.Println("Error parsing form data:", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	validator := &validator.Validator{
		Errors: make(map[string]string),
	}

	email := r.FormValue("email")
	validator.CheckField(validator.NotBlank(email), "email", "Please, fill the email field")
	if validator.Errors["email"] == "" {
		validator.CheckField(validator.ValidEmail(email), "email", "Invalid email")
	}

	password := r.FormValue("password")
	validator.CheckField(validator.NotBlank(password), "password", "Please, fill the password field")

	if !validator.Valid() {
		app.sendJSONResponse(w, 200, validator)
		return
	}

	uid := app.user.EmailExist(email)
	if uid == 0 {
		resp := app.sendMessage(false, "Incorrect Credentials")
		app.sendJSONResponse(w, 200, resp)
		return
	}

	uid, err = app.user.Authenticate(email, password)
	if err != nil {
		resp := app.sendMessage(false, err.Error())
		app.sendJSONResponse(w, 200, resp)
		return
	}

	if uid == 0 {
		resp := app.sendMessage(false, "Incorrect Credentials")
		app.sendJSONResponse(w, 200, resp)
		return
	}

	hashed := app.generateHash(r.RemoteAddr, r.URL.Port())
	err = app.user.CreateAuthHeader(hashed, uid)
	if err != nil {
		app.serverError(w, err)
		return
	}

	// Set bearer token in header and send response
	w.Header().Set("Authorization", "Bearer "+hashed)
	resp := app.sendMessage(true, "Login Successfull")
	app.sendJSONResponse(w, 200, resp)
}

func (app *application) UserLogout(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	token := app.user.CheckBearerToken(authHeader)
	if token != "" {
		err := app.user.Logout(token)
		if err != nil {
			app.serverError(w, err)
			return
		}
	}

	resp := app.sendMessage(true, "Logout Successfull")
	app.sendJSONResponse(w, 200, resp)
}

func (app *application) NewPasswordPost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.errorLog.Println("Error parsing form data:", err)
		app.CustomError(w, "Bad request", http.StatusBadRequest)
		return
	}

	validator := &validator.Validator{
		Errors: make(map[string]string),
	}
	password := r.FormValue("password")
	repeatPassword := r.FormValue("repeatPassword")
	validator.CheckField(validator.NotBlank(password), "password", "Please, fill the password field")
	if validator.Errors["password"] == "" {
		validator.CheckField(validator.MaxChars(password, 15), "password", "Password should not be less than than 15 character")
	}

	validator.CheckField(validator.NotBlank(repeatPassword), "repeatPassword", "Empty Repeat Password")
	if validator.Errors["repeatPassword"] == "" && validator.Errors["password"] == "" {
		if password != repeatPassword {
			validator.Errors["repeatPassword"] = "Password not matched"
		}
	}

	if !validator.Valid() {
		app.sendJSONResponse(w, 200, validator)
		return
	}
	params := httprouter.ParamsFromContext(r.Context())
	uri := params.ByName("uri")

	uid, err := app.user.ForgetPasswordUri(uri)
	if err != nil {
		app.notFound(w)
		return
	}

	err = app.user.NewPassword(password, uid)
	if err != nil {
		app.CustomError(w, "Internal Server Error"+err.Error(), 500)
		return
	}

	resp := app.sendMessage(true, "Password Changed Successfully")
	app.sendJSONResponse(w, 200, resp)
}

func (app *application) ForgetPasswordPost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.errorLog.Println("Error parsing form data:", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	validator := &validator.Validator{
		Errors: make(map[string]string),
	}
	email := r.FormValue("email")
	validator.CheckField(validator.NotBlank(email), "email", "Please, fill the email field")
	if validator.Errors["email"] == "" {
		validator.CheckField(validator.ValidEmail(email), "email", "Invalid email")
	}

	if !validator.Valid() {
		app.sendJSONResponse(w, 200, validator)
		return
	}

	uid := app.user.EmailExist(email)
	if uid > 0 {
		token := fmt.Sprintf("%s_%d", email, uid)
		uri := app.generateHash(token, r.RemoteAddr)
		err := app.user.ForgetPassword(uid, uri)
		if err != nil {
			app.CustomError(w, "Internal Server Error", 400)
		}
	}

	// whether email is registered or not we will show same success message
	resp := app.sendMessage(true, "If email is registered email you will get link on your email")
	app.sendJSONResponse(w, 200, resp)
}

// create  a token for login, user_verification
func (app *application) generateHash(addr, port string) string {
	data := addr + port + strconv.FormatInt(time.Now().Unix(), 10)
	hasher := sha1.New()
	hasher.Write([]byte(data))
	hash := hasher.Sum(nil)
	// Convert hash bytes to a hexadecimal string
	hashStr := fmt.Sprintf("%x", hash)

	return hashStr
}

func (app *application) sendJSONResponse(w http.ResponseWriter, statusCode int, message any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(message)
}

func (app *application) sendMessage(status bool, message string) Message {
	return Message{
		Status:  status,
		Message: message,
	}
}

func (app *application) ValidToken(w http.ResponseWriter, r *http.Request) int {
	authHeader := r.Header.Get("Authorization")
	token := app.user.CheckBearerToken(authHeader)
	if token == "" {
		return 0
	}

	uid := app.user.ValidToken(token)
	return uid
}

package main

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"net/http"
	// "regexp"
	"strconv"

	"github.com/julienschmidt/httprouter"
	"test.iamgak.net/validator"

	// "strings"
	"time"
)

type Message struct {
	Status  any    `json:"status"`
	Message string `json:"message"`
}

func (app *application) Home(w http.ResponseWriter, r *http.Request) {
	resp := Message{
		Status:  true,
		Message: "Welcome to our Bookstore, Website",
	}

	app.sendJSONResponse(w, 200, &resp)
}

// book related handlers
// listing
func (app *application) BooksListing(w http.ResponseWriter, r *http.Request) {
	bks, err := app.books.BooksListing()
	if err != nil {
		app.errorLog.Print("internal server error")
		app.CustomError(w, "Internal Server Error", 500)
		return
	}

	app.sendJSONResponse(w, 200, bks)
}

func (app *application) BooksInfo(w http.ResponseWriter, r *http.Request) {
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

func (app *application) AddBooks(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	token := app.user.CheckBearerToken(authHeader)
	if token == "" {
		app.CustomError(w, "Authorization failed. Please provide a valid bearer token to access this resource", 401)
		return
	}

	uid := app.user.ValidToken(token)
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
		validator.CheckField(validator.MaxChars(isbn, 100), "isbn", "Please, fill the ISBN shorter than 20")
	}

	if validator.Errors["author"] == "" {
		validator.CheckField(validator.MaxChars(isbn, 20), "author", "Please, fill the AUTHOR shorter than 20")
	}
	if validator.Errors["title"] == "" {
		validator.CheckField(validator.MaxChars(isbn, 20), "title", "Please, fill the TITLE shorter than 20")
	}

	price, err := strconv.ParseFloat(r.FormValue("price"), 32)
	if err != nil {
		validator.Errors["price"] = "Invalid price value"
	}

	if validator.Valid() {
		book_id, _ := app.books.BookExist(isbn)
		if book_id != 0 {
			validator.Errors["isbn"] = "isbn already Exist"
		}
	}

	if !validator.Valid() {
		app.sendJSONResponse(w, 200, *validator)
		return
	}

	_, err = app.books.InsertBook(isbn, author, title, genre, descriptions, price, uid)
	if err != nil {
		app.CustomError(w, "Server Issue", 500)
		return
	}

	resp := Message{
		Status:  true,
		Message: "Book Review Saved!!!",
	}

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
	app.infoLog.Print(repeatPassword)
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
	} else if app.user.EmailExist(email) == 0 {
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

	hashed := app.generateHash(r.RemoteAddr, r.URL.Port())
	exist, err := app.user.InsertUser(email, password, hashed)
	if err != nil || !exist {
		app.serverError(w, err)
		return
	}

	resp := Message{
		Status:  true,
		Message: "Registration Successfull " + email,
	}

	app.sendJSONResponse(w, 200, resp)
}

func (app *application) UserActivation(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	uri := params.ByName("uri")
	err := app.user.AccountActivate(uri)
	if err != nil {
		app.infoLog.Print("hello")
		app.notFound(w)
		return
	}

	fmt.Fprintln(w, "your account has been verified...")
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

	// email, password := r.FormValue("email"), r.FormValue("password")
	uid := app.user.Authenticate(email, password)
	if uid == 0 {
		resp := app.sendMessage(false, "Incorrect Credentials")
		app.sendJSONResponse(w, 200, resp)
		return
	}

	hashed := app.generateHash(r.RemoteAddr, r.URL.Port())
	err = app.user.Authorization(hashed, uid)
	if err != nil {
		app.serverError(w, err)
		return
	}

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
		http.Error(w, "Bad request", http.StatusBadRequest)
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
		rows, err := app.user.ForgetPassword(uid, uri)
		if err != nil || rows == 0 {
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

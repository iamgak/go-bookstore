package main

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"net/http"
	// "regexp"
	"strconv"
	"time"
)

type Message struct {
	Status  any    `json:"status"`
	Message string `json:"message"`
}

func (app *application) Home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		app.notFound(w)
		return
	}

	if r.Method == "GET" {
		resp := Message{
			Status:  true,
			Message: "Welcome to our Bookstore, Website",
		}

		app.sendJSONResponse(w, 200, &resp)
		return
	}

	err := r.ParseForm()
	if err != nil {
		app.errorLog.Println("Error parsing form data:", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}
	validator := app.validator

	email := r.FormValue("email")
	password := r.FormValue("password")
	repeatPassword := r.FormValue("repeatPassword")

	validator.CheckField(validator.NotBlank(email), "email", "Please, fill the email field")
	if validator.Errors["email"] == "" {
		validator.CheckField(validator.ValidEmail(email), "email", "Invalid email")
	}

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

	resp := Message{
		Status:  true,
		Message: "Submitted Successfully",
	}
	app.sendJSONResponse(w, 200, resp)
}

// book related handlers --done
func (app *application) BooksListing(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		app.CustomError(w, "METHOD NOT ALLOWED", 405)
		return
	}

	bks, err := app.books.BooksListing()
	if err != nil {
		app.errorLog.Print("internal server error")
		app.CustomError(w, "Internal Server Error", 500)
		return
	}

	app.sendJSONResponse(w, 200, bks)
}

// book-info
func (app *application) BooksInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		app.CustomError(w, "METHOD NOT ALLOWED", 405)
	}

	isbn := r.URL.Query().Get("isbn")
	validator := app.validator
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
	if r.Method != "POST" {
		app.CustomError(w, "Method Not allowed", 403)
		return
	}

	authHeader := r.Header.Get("Authorization")
	// Check if token is valid
	token := app.user.CheckBearerToken(authHeader)
	if token == "" {
		app.CustomError(w, "Authorization failed. Please provide a valid bearer token to access this resource", 401)
		return
	}

	uid, _ := app.user.ValidToken(token)
	if uid == 0 {
		app.CustomError(w, "Authorization failed. Please provide a valid bearer token to access this resource", 401)
		return
	}

	isbn := r.FormValue("isbn")
	title := r.FormValue("title")
	author := r.FormValue("author")
	genre := r.FormValue("genre")
	descriptions := r.FormValue("descriptions")
	validator := app.validator
	validator.CheckField(validator.NotBlank(isbn), "ISBN", "Please, fill the ISBN field")
	validator.CheckField(validator.NotBlank(author), "AUTHOR", "Please, fill the Author field")
	validator.CheckField(validator.NotBlank(title), "TITLE", "Please, fill the title field")

	if validator.Errors["ISBN"] == "" {
		validator.CheckField(validator.MaxChars(isbn, 20), "ISBN", "Please, fill the ISBN shorter than 20")
	}

	if validator.Errors["AUTHOR"] == "" {
		validator.CheckField(validator.MaxChars(isbn, 20), "AUTHOR", "Please, fill the AUTHOR shorter than 20")
	}
	if validator.Errors["TITLE"] == "" {
		validator.CheckField(validator.MaxChars(isbn, 20), "TITLE", "Please, fill the TITLE shorter than 20")
	}

	price, err := strconv.ParseFloat(r.FormValue("price"), 32)
	if err != nil {
		validator.Errors["PRICE"] = "Invalid price value"
	}

	if validator.Valid() {
		book_id, _ := app.books.BookExist(isbn)
		if book_id != 0 {
			validator.Errors["ISBN"] = "ISBN already Exist"
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

// user related handlers
func (app *application) UserRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		app.CustomError(w, "Method Not Allowd", 405)
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")
	validator := app.validator
	validator.CheckField(validator.NotBlank(email), "email", "Please, fill the email field")
	validator.CheckField(validator.NotBlank(password), "password", "Please, fill the password field")

	var exist bool
	exist, err := app.user.EmailExist(email)
	if err != nil {
		app.errorLog.Print(w, err)
		app.CustomError(w, "Internal Server Error122", 500)
		return
	}

	if !exist {
		validator.CheckField(validator.NotBlank(email), "ISBN", "Email already registered")
	}

	if !validator.Valid() {
		app.sendJSONResponse(w, 200, validator)
		return
	}

	hashed := app.generateHash(r.RemoteAddr, r.URL.Port())
	exist, err = app.user.InsertUser(email, password, hashed)
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
	token := r.URL.Query().Get("token")
	if r.Method == "POST" || token == "" {
		app.notFound(w)
		return
	}

	err := app.user.AccountActivate(token)
	if err != nil {
		app.notFound(w)
		return
	}

	fmt.Fprintln(w, "your account has been verified...")
}

func (app *application) UserLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		app.notFound(w)
		return
	}

	email, password := r.FormValue("email"), r.FormValue("password")
	var uid int64
	uid, err := app.user.Authenticate(email, password)
	if err != nil || uid == 0 {
		resp := Message{
			Status:  true,
			Message: "Incorrect Credentials",
		}
		app.sendJSONResponse(w, 200, resp)
		return
	}

	hashed := app.generateHash(r.RemoteAddr, r.URL.Port())
	err = app.user.Authorization(hashed, uid)
	if err != nil {
		app.serverError(w, err)
		return
	}

	resp := Message{
		Status:  true,
		Message: "Login Successfull",
	}

	app.sendJSONResponse(w, 200, resp)
}

func (app *application) UserLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		app.CustomError(w, "Method Not allowed", 403)
		return
	}

	authHeader := r.Header.Get("Authorization")
	token := app.user.CheckBearerToken(authHeader)
	if token == "" {
		app.CustomError(w, "Authorization failed. Please provide a valid bearer token to access this resource", 401)
		return
	}

	err := app.user.Logout(token)
	if err != nil {
		app.serverError(w, err)
		return
	}

	fmt.Println("User logout Succefully")
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

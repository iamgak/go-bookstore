package main

import (
	// "Errors"
	"crypto/sha1"
	// "database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	// "regexp"
	"strconv"
	"strings"

	// "test.iamgak.net/validator"
	"time"

	// "test.iamgak.net/models"
	"test.iamgak.net/validator"
	// "test.iamgak.net/models"
)

type Credentials struct {
	Email          string `json:"email"`
	Password       string `json:"password"`
	RepeatPassword string `json:"repeatpassword"`
	// Password string `json:"password"`
}

type LoginResponse struct {
	Email string `json:"email"`
	Token string `json:"token"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type CreateAccountRequest struct {
	Email             string `json:"email"`
	Password          string `json:"password"`
	VerificationToken string `json:"token"`
}

type Message struct {
	Status  any    `json:"status"`
	Message string `json:"message"`
}

type Response struct {
	Message map[string]string
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

	var credentials = Credentials{
		Email:          r.Form.Get("email"),
		Password:       r.Form.Get("password"),
		RepeatPassword: r.Form.Get("repeatPassword"),
	}

	validator := &validator.Validator{
		Errors: make(map[string]string),
	}

	validator.CheckField(validator.NotBlank(credentials.Email), "email", "Please, fill the email field")
	if validator.Errors["email"] == "" {
		validator.CheckField(validator.ValidEmail(credentials.Email), "email", "Invalid email")
	}

	validator.CheckField(validator.NotBlank(credentials.Password), "password", "Please, fill the password field")
	if validator.Errors["password"] == "" {
		validator.CheckField(validator.MaxChars(credentials.Password, 15), "password", "Password should not be less than than 15 character")
	}

	validator.CheckField(validator.NotBlank(credentials.RepeatPassword), "repeatPassword", "Empty Repeat Password")
	if validator.Errors["repeatPassword"] == "" && validator.Errors["password"] == "" {
		if credentials.Password != credentials.RepeatPassword {
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

// book related handlers
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

func (app *application) BooksInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		app.CustomError(w, "METHOD NOT ALLOWED", 405)
	}

	isbn := r.URL.Query().Get("isbn")
	validator := &validator.Validator{}
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
	validator := &validator.Validator{
		Errors: make(map[string]string),
	}
	validator.CheckField(validator.NotBlank(authHeader), "Token", "No Bearer Token")
	if !validator.Valid() {
		app.sendJSONResponse(w, 200, validator.Errors)
		return
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")

	// Check if token is valid
	isValid, err := app.user.CheckBearerToken(token)
	if err != nil {
		app.serverError(w, err)
		return
	}

	if !isValid {
		validator.Errors["Authorization"] = "false"
		app.sendJSONResponse(w, 200, validator.Errors)
		return
	}

	isbn := r.FormValue("isbn")
	title := r.FormValue("title")
	author := r.FormValue("author")

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

	_, err = app.books.InsertBook(isbn, author, title, price)
	if err != nil {
		app.CustomError(w, "Server Issue", 500)
		return
	}

	response := &Response{
		Message: make(map[string]string),
	}

	response.Message["success"] = "true"
	response.Message["Message"] = "Book saved successfully"
	app.sendJSONResponse(w, 200, *response)

}

// user related handlers
func (app *application) UserRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		app.CustomError(w, "Method Not Allowd", 405)
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")
	validator := &validator.Validator{
		Errors: make(map[string]string),
	}

	validator.CheckField(validator.NotBlank(email), "email", "Please, fill the email field")
	validator.CheckField(validator.NotBlank(password), "password", "Please, fill the password field")

	var uid int64
	uid, err := app.user.EmailExist(email)
	if err != nil {
		app.errorLog.Print(w, err)
		app.CustomError(w, "Internal Server Error122", 500)
		return
	}

	if uid != 0 {
		validator.CheckField(validator.NotBlank(email), "ISBN", "Email already registered")
	}

	if !validator.Valid() {
		app.sendJSONResponse(w, 200, validator)
		return
	}

	hashed := app.generateHash(r.RemoteAddr, r.URL.Port())

	// var data CreateAccountRequest
	data := CreateAccountRequest{
		Email:             email,
		Password:          password,
		VerificationToken: hashed,
	}

	uid, err = app.user.InsertUser(&data)
	if err != nil || uid == 0 {
		app.serverError(w, err)
		return
	}

	response := &Response{
		Message: make(map[string]string),
	}

	response.Message["status"] = "true"
	response.Message["message"] = "Registration successfull"
	response.Message["hashed"] = hashed
	app.sendJSONResponse(w, 200, response)
}

func (app *application) UserActivation(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if r.Method == "POST" || token == "" {
		app.notFound(w)
		return
	}

	result, err := app.db.Exec("UPDATE users SET `verified` = NULL, `active` = 1 WHERE verified = ? ", token)
	if err != nil {
		app.serverError(w, err)
		return
	}

	rowAffected, _ := result.RowsAffected()
	if rowAffected == 0 {
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

	req := LoginRequest{
		Email:    r.FormValue("email"),
		Password: r.FormValue("password"),
	}

	app.sendJSONResponse(w, 200, req)
	var uid int64
	uid, err := app.user.Authenticate(&req)
	if err != nil || uid == 0 {
		http.Error(w, "Incorrect credentials ", 500)
		return
	}

	hashed := app.generateHash(r.RemoteAddr, r.URL.Port())

	result, err := app.db.Exec("UPDATE users SET `token` =? WHERE  `id` =  ?", hashed, uid)
	if err != nil {
		app.serverError(w, err)
		return
	}

	rowAffected, _ := result.RowsAffected()
	if rowAffected == 0 {
		app.serverError(w, err)

		return
	}

	fmt.Fprintf(w, "your Login bearer token: %s", hashed)
}

func (app *application) UserLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		app.CustomError(w, "Method Not allowed", 403)
		return
	}

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		app.CustomError(w, "Empty Authorization failed.", 401)
		return
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")

	// Check if token is valid
	isValid, err := app.user.CheckBearerToken(token)
	if err != nil {
		app.serverError(w, err)
		return
	}

	if !isValid {
		app.CustomError(w, "Authorization failed. Please provide a valid bearer token to access this resource", 401)
		return
	}

	_, err = app.db.Exec("UPDATE users SET `token` = NULL WHERE  `token` =  ?", token)
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

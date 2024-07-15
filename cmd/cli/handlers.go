package main

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"
	"test.iamgak.net/models"
	"test.iamgak.net/validator"
)

type Message struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
}

// home page
func (app *application) Home(w http.ResponseWriter, r *http.Request) {
	resp := Message{
		Status:  true,
		Message: "Welcome to our Bookstore, Website",
	}

	app.sendJSONResponse(w, 200, &resp)
}

// all the review listing
func (app *application) ReviewListing(w http.ResponseWriter, r *http.Request) {
	bks, err := app.review.ReviewListing()
	if err != nil {
		app.errorLog.Print("internal server error")
		app.CustomError(w, "Internal Server Error", 500)
		return
	}

	app.sendJSONResponse(w, 200, bks)
}

// logged user review
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

// search by isbn only
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

	app.user.ActivityLog("review_deleted", uid)
	resp := app.sendMessage(true, "Review Deleted")
	app.sendJSONResponse(w, 200, resp)
}

// create new review

func (app *application) AddReview(w http.ResponseWriter, r *http.Request) {
	uid := app.ValidToken(w, r)
	if uid == 0 {
		app.CustomError(w, "Authorization failed. Please provide a valid bearer token to access this resource", 401)
		return
	}

	var CreateReview *models.Review
	err := json.NewDecoder(r.Body).Decode(&CreateReview)
	if err != nil {
		app.errorLog.Print(err)
		app.CustomError(w, "header should be application/json and all the field should be in proper format", 400)
		return
	}

	CreateReview.Uid = uid
	validator := &validator.Validator{
		Errors: make(map[string]string),
	}

	validator.CheckField(validator.NotBlank(CreateReview.Isbn), "isbn", "Please, fill the isbn field")
	validator.CheckField(validator.NotBlank(CreateReview.Descriptions), "descriptions", "Please, fill the descriptions field")
	validator.CheckField(validator.NotBlank(CreateReview.Title), "title", "Please, fill the title field")

	if validator.Errors["isbn"] == "" {
		validator.CheckField(validator.MaxChars(CreateReview.Title, 20), "isbn", "Please, fill the ISBN shorter than 20")
	}

	if validator.Errors["descriptions"] == "" {
		validator.CheckField(validator.MaxChars(CreateReview.Descriptions, 100), "isbn", "Please, fill the ISBN shorter than 100")
	}

	if validator.Errors["title"] == "" {
		validator.CheckField(validator.MaxChars(CreateReview.Title, 50), "title", "Please, fill the TITLE shorter than 50")
	}

	if validator.Valid() {
		book_id := app.books.BookExist(CreateReview.Isbn)
		if book_id {
			validator.Errors["isbn"] = "isbn already Exist"
		}
	}

	if !validator.Valid() {
		app.sendJSONResponse(w, 200, validator)
		return
	}

	err = app.review.CreateReview(CreateReview)
	if err != nil {
		app.CustomError(w, "Server Issue", 500)
		return
	}

	app.user.ActivityLog("review_created", uid)
	resp := app.sendMessage(true, "Review Saved")
	app.sendJSONResponse(w, 200, resp)
}

// bookListing related handlers
func (app *application) BookListing(w http.ResponseWriter, r *http.Request) {

	bks, err := app.books.BooksListing()
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.sendJSONResponse(w, 200, bks)
}

// book info based on isbn
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

	info, err := app.books.GetBookByIsbn(isbn)
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.sendJSONResponse(w, 200, info)
}

// add book in db
func (app *application) AddBook(w http.ResponseWriter, r *http.Request) {
	uid := app.ValidToken(w, r)
	if uid == 0 {
		app.CustomError(w, "Authorization failed. Please provide a valid bearer token to access this resource", 401)
		return
	}

	var bookRegister *models.Book
	err := json.NewDecoder(r.Body).Decode(&bookRegister)
	if err != nil {
		app.errorLog.Print(err)
		app.CustomError(w, "Unsupported or empty fields", 400)
		return
	}

	validator := &validator.Validator{
		Errors: make(map[string]string),
	}

	validator.CheckField(validator.NotBlank(bookRegister.ISBN), "isbn", "Please, fill the isbn field")
	validator.CheckField(validator.NotBlank(bookRegister.Genre), "genre", "Please, fill the genre field")
	validator.CheckField(validator.NotBlank(bookRegister.Descriptions), "descriptions", "Please, fill the descriptions field")
	validator.CheckField(validator.NotBlank(bookRegister.Author), "author", "Please, fill the author field")
	validator.CheckField(validator.NotBlank(bookRegister.Title), "title", "Please, fill the title field")

	if validator.Errors["isbn"] == "" {
		validator.CheckField(validator.MaxChars(bookRegister.Title, 20), "isbn", "Please, fill the ISBN shorter than 20")
	}

	if validator.Errors["descriptions"] == "" {
		validator.CheckField(validator.MaxChars(bookRegister.Descriptions, 100), "isbn", "Please, fill the ISBN shorter than 100")
	}

	if validator.Errors["author"] == "" {
		validator.CheckField(validator.MaxChars(bookRegister.Author, 50), "author", "Please, fill the AUTHOR shorter than 50")
	}
	if validator.Errors["title"] == "" {
		validator.CheckField(validator.MaxChars(bookRegister.Title, 50), "title", "Please, fill the TITLE shorter than 50")
	}

	if validator.Valid() {
		book_id := app.books.BookExist(bookRegister.ISBN)
		if book_id {
			validator.Errors["isbn"] = "isbn already Exist"
		}
	}

	if !validator.Valid() {
		app.sendJSONResponse(w, 200, validator)
		return
	}

	err = app.books.CreateBook(bookRegister)
	if err != nil {
		app.CustomError(w, "Server Issue1", 500)
		return
	}

	app.user.ActivityLog("Book Listed", uid)
	resp := app.sendMessage(true, "Book Record Saved, Sucessfully")
	app.sendJSONResponse(w, 200, resp)
}

// register user
func (app *application) UserRegister(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.errorLog.Println("Error parsing form data:", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	var creds *models.UserRegister
	err = json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		app.errorLog.Print(err)
		app.CustomError(w, "Unsupported or empty fields", 400)
		return
	}

	validator := &validator.Validator{
		Errors: make(map[string]string),
	}

	validator.CheckField(validator.NotBlank(creds.Email), "email", "Please, fill the email field")
	validator.CheckField(validator.NotBlank(creds.Password), "password", "Please, fill the password field")
	validator.CheckField(validator.NotBlank(creds.RepeatPassword), "repeatPassword", "Please, fill the repeat password field")

	validator.ValidPassword(creds.Password)

	if validator.Errors["email"] == "" {
		validator.CheckField(validator.ValidEmail(creds.Email), "email", "Invalid Email Format")
	}
	if app.user.EmailExist(creds.Email) != 0 {
		validator.CheckField(false, "email", "Email already registered")
	}

	if validator.Errors["repeatPassword"] == "" && validator.Errors["password"] == "" {
		if creds.Password != creds.RepeatPassword {
			validator.Errors["repeatPassword"] = "Password not matched"
		}
	}

	if !validator.Valid() {
		app.sendJSONResponse(w, 200, validator)
		return
	}

	uri := app.generateHash(r.RemoteAddr, r.URL.Port())
	uid, err := app.user.InsertUser(creds.Email, creds.Password, uri)
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.user.ActivityLog("Account Created", uid)
	resp := Message{
		Status:  true,
		Message: "Registration Successfull ",
	}

	app.sendJSONResponse(w, 200, resp)
}

// after registration user need to validate using link save in users table respective email
func (app *application) UserActivation(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	uri := params.ByName("uri")
	if !app.user.ValidURI(uri) {
		app.notFound(w)
		return
	}

	err := app.user.AccountActivate(uri)
	if err != nil {
		app.notFound(w)
		return
	}

	resp := app.sendMessage(true, "Your account has been Verified.")
	app.sendJSONResponse(w, 200, resp)
}

// login
func (app *application) UserLogin(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.errorLog.Println("Error parsing form data:", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	var creds *models.UserLogin
	err = json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		app.errorLog.Print(err)
		app.CustomError(w, "Unsupported or empty fields", 400)
		return
	}

	validator := &validator.Validator{
		Errors: make(map[string]string),
	}

	validator.CheckField(validator.NotBlank(creds.Email), "email", "Please, fill the email field")
	validator.CheckField(validator.NotBlank(creds.Password), "password", "Please, fill the password field")

	if validator.Errors["email"] == "" {
		validator.CheckField(validator.ValidEmail(creds.Email), "email", "Invalid Email Format")
	}

	if !validator.Valid() {
		app.sendJSONResponse(w, 200, validator)
		return
	}

	uid := app.user.EmailExist(creds.Email)
	if uid == 0 {
		resp := app.sendMessage(false, "Incorrect Credentials")
		app.sendJSONResponse(w, 200, resp)
		return
	}

	uid, err = app.user.Login(creds)
	if err != nil {
		app.CustomError(w, "Internal Server Error", 500)
		return
	}

	if uid == 0 {
		resp := app.sendMessage(false, "Incorrect Credentials")
		app.sendJSONResponse(w, 200, resp)
		return
	}

	hashed := app.generateHash(r.RemoteAddr, r.URL.Port())
	err = app.user.SetLoginToken(hashed, uid)
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.user.ActivityLog("logged_in", uid)
	// Set bearer token in header and send response
	w.Header().Set("Authorization", "Bearer "+hashed)
	resp := app.sendMessage(true, "Login Successfull")
	app.sendJSONResponse(w, 200, resp)
}

func (app *application) UserLogout(w http.ResponseWriter, r *http.Request) {
	uid := app.ValidToken(w, r)
	if uid != 0 {
		err := app.user.Logout(uid)
		if err != nil {
			app.serverError(w, err)
			return
		}

		app.user.ActivityLog("log_out", uid)
		resp := app.sendMessage(true, "Logout Successfull")
		app.sendJSONResponse(w, 200, resp)
	}

	app.notFound(w)
}

// after forget password it create uri in db like /new_password/db_uri
func (app *application) NewPasswordPost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.errorLog.Println("Error parsing form data:", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	params := httprouter.ParamsFromContext(r.Context())
	uri := params.ByName("uri")

	uid, err := app.user.ForgetPasswordUri(uri)
	if err != nil {
		app.notFound(w)
		return
	}

	var creds *models.UserNewPassword
	err = json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		app.errorLog.Print(err)
		app.CustomError(w, "Unsupported or empty fields", 400)
		return
	}

	validator := &validator.Validator{
		Errors: make(map[string]string),
	}

	validator.CheckField(validator.NotBlank(creds.Password), "password", "Please, fill the password field")

	validator.CheckField(validator.NotBlank(creds.RepeatPassword), "repeatPassword", "Please, fill the password field")
	if validator.Errors["repeatPassword"] == "" && validator.Errors["password"] == "" {
		if creds.Password != creds.RepeatPassword {
			validator.Errors["repeatPassword"] = "Password not matched"
		}
	}

	if !validator.Valid() {
		app.sendJSONResponse(w, 200, validator)
		return
	}

	err = app.user.NewPassword(creds.Password, uid)
	if err != nil {
		app.CustomError(w, "Internal Server Error"+err.Error(), 500)
		return
	}

	resp := app.sendMessage(true, "Password Changed Successfully")
	app.sendJSONResponse(w, 200, resp)
}

// it will create a uri if email is valid and active user
func (app *application) ForgetPasswordPost(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.errorLog.Println("Error parsing form data:", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	var creds *models.ForgetPassword
	err = json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		app.errorLog.Print(err)
		app.CustomError(w, "Empty field / Unsupported Content-Type", 400)
		return
	}

	validator := &validator.Validator{
		Errors: make(map[string]string),
	}

	validator.CheckField(validator.NotBlank(creds.Email), "email", "Please, fill the email field")

	if !validator.Valid() {
		app.sendJSONResponse(w, 200, validator)
		return
	}

	uid := app.user.EmailExist(creds.Email)
	if uid > 0 {
		token := app.generateHash(creds.Email, r.RemoteAddr)
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

// create a token for login, user_verification
func (app *application) generateHash(addr, port string) string {
	data := addr + port + strconv.FormatInt(time.Now().Unix(), 10)
	hasher := sha1.New()
	hasher.Write([]byte(data))
	hash := hasher.Sum(nil)
	// Convert hash bytes to a hexadecimal string
	hashStr := fmt.Sprintf("%x", hash)

	return hashStr
}

// to print json message
func (app *application) sendJSONResponse(w http.ResponseWriter, statusCode int, message any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(message)
}

// create a struct Message to send data in sendJSONResponse
func (app *application) sendMessage(status bool, message string) Message {
	return Message{
		Status:  status,
		Message: message,
	}
}

func (app *application) ValidToken(w http.ResponseWriter, r *http.Request) int64 {
	authHeader := r.Header.Get("Authorization")
	token := app.user.CheckBearerToken(authHeader)
	if token == "" {
		return 0
	}

	uid := app.user.ValidToken(token)
	return uid
}

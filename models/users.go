package models

import (
	"database/sql"
)

type LoginResponse struct {
	Email string `json:"email"`
	Token string `json:"token"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type CreateAccountRequest struct {
	Email    string `json:"email"`
	Verified string `json:"token"`
	Password string `json:"password"`
}

// Define a new User type. Notice how the field names and types align
// with the columns in the database "users" table?
type User struct {
	Id    int
	Email string
}

// Define a new UserModel type which wraps a database connection pool.
type UserModel struct {
	DB *sql.DB
}

// We'll use the Insert method to add a new record to the "users" table.
func (m *UserModel) InsertUser(userInfo *CreateAccountRequest) (int64, error) {
	result, err := m.DB.Exec("INSERT INTO users(`email`,`password`,`verified`) VALUES (?, password(?),? )", userInfo.Email, userInfo.Password, userInfo.Verified)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

// We'll use the Authenticate method to verify whether a user exists with
// the provided email address and password. This will return the relevant
// user ID if they do.
func (m *UserModel) Authenticate(user *LoginRequest) (int64, error) {
	var uid int64
	err := m.DB.QueryRow("SELECT id FROM `users` WHERE `active` = 1 AND `email` = ? AND `password` = PASSWORD(?) ", user.Email, user.Password).Scan(&uid)
	if err != nil {
		return 0, err
	}

	if uid == 0 {
		return 0, nil
	}

	return uid, nil
}

func (m *UserModel) EmailExist(email string) (int64, error) {
	var valid int64
	err := m.DB.QueryRow("SELECT 1 FROM `users` WHERE  `email` = ?", email).Scan(&valid)
	if err != nil && err != sql.ErrNoRows {
		return 0, err
	}

	if valid == 0 {
		return 0, nil
	}

	return valid, nil
}

func (m *UserModel) ValidToken(token string) (*User, error) {
	user := &User{}
	err := m.DB.QueryRow("SELECT email, id FROM users WHERE token = ? ", token).Scan(&user.Email, &user.Id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNoRecord
		} else {
			return nil, err
		}
	}

	return user, nil
}

func (m *UserModel) CheckBearerToken(token string) (bool, error) {

	var count int
	err := m.DB.QueryRow("SELECT COUNT(*) FROM users WHERE token = ? ", token).Scan(&count)
	if err != nil {
		return false, err
	}

	// If count is greater than 0, token is valid
	return count > 0, nil
}

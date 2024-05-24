package models

import (
	"database/sql"
	"strings"
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

// type UserModels interface {
// 	InsertUser(string, string, string) (bool, error)
// 	Authenticate(string, string) (int64, error)
// 	EmailExist(string) (bool, error)
// }

// Define a new UserModel type which wraps a database connection pool.
type UserModel struct {
	DB *sql.DB
}

// We'll use the Insert method to add a new record to the "users" table.
func (m *UserModel) InsertUser(email, password, hashed string) (bool, error) {
	_, err := m.DB.Exec("INSERT INTO users(`email`,`password`,`verified`) VALUES (?, password(?),? )", email, password, hashed)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (m *UserModel) Authorization(token string, uid int64) error {

	_, err := m.DB.Exec("UPDATE `users` SET `token` = ? WHERE `id` = ?", token, uid)
	if err != nil {
		return err
	}
	return nil
}

func (m *UserModel) Logout(token string) error {

	_, err := m.DB.Exec("UPDATE `users` SET `token` = NULL WHERE `token` = ?", token)
	if err != nil {
		return err
	}
	return nil
}
func (m *UserModel) Authenticate(email, password string) (int64, error) {
	var uid int64
	err := m.DB.QueryRow("SELECT id FROM `users` WHERE `active` = 1 AND `email` = ? AND `password` = PASSWORD(?) ", email, password).Scan(&uid)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}

		return 0, err
	}

	return uid, nil
}

func (m *UserModel) EmailExist(email string) (bool, error) {
	var valid int64
	err := m.DB.QueryRow("SELECT 1 FROM `users` WHERE  `email` = ?", email).Scan(&valid)
	if err != nil && err != sql.ErrNoRows {
		return false, err
	}

	if valid == 0 {
		return false, nil
	}

	return true, nil
}

func (m *UserModel) ValidToken(token string) (int, error) {

	var id int
	err := m.DB.QueryRow("SELECT  id FROM users WHERE token = ? ", token).Scan(id)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, ErrNoRecord
		} else {
			return 0, err
		}
	}

	return id, nil
}

func (m *UserModel) AccountActivate(token string) error {

	result, err := m.DB.Exec("UPDATE users SET `verified` = NULL, `active` = 1 WHERE verified = ? ", token)
	if err != nil {
		return err
	}

	_, err = result.RowsAffected()
	return err
}
func (m *UserModel) CheckBearerToken(token string) string {

	if token == "" {
		return ""
	}

	return strings.TrimPrefix(token, "Bearer ")

}

// var _ UserModels = (*UserModel)(nil)

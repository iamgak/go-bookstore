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

type User struct {
	Id    int
	Email string
}

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
func (m *UserModel) Authenticate(email, password string) int64 {
	var uid int64
	_ = m.DB.QueryRow("SELECT id FROM `users` WHERE `active` = 1 AND `email` = ? AND `password` = PASSWORD(?) ", email, password).Scan(&uid)

	return uid
}

func (m *UserModel) EmailExist(email string) int64 {
	var uid int64
	_ = m.DB.QueryRow("SELECT id FROM `users` WHERE  `email` = ?", email).Scan(&uid)
	return uid
}

func (m *UserModel) ValidToken(token string) int {
	var id int
	_ = m.DB.QueryRow("SELECT id FROM users WHERE token = ? ", token).Scan(&id)
	return id
}

func (m *UserModel) AccountActivate(token string) error {
	result, err := m.DB.Exec("UPDATE users SET `verified` = NULL, `active` = 1 WHERE verified = ? ", token)
	if err != nil {
		return err
	}

	_, err = result.RowsAffected()
	return err
}

func (m *UserModel) ForgetPassword(uid int64, uri string) (int64, error) {
	_, _ = m.DB.Exec("UPDATE `forget_passw` SET superseded = 1 WHERE uid = ?", uid)
	data, err := m.DB.Exec("INSERT INTO `forget_passw` (uid,uri,superseded) VALUES(?,?,0) ", uid, uri)
	if err != nil {
		return 0, err
	}

	rowsAffected, err := data.RowsAffected()
	return rowsAffected, nil
}

func (m *UserModel) ForgetPasswordUri(uri string) (int, error) {
	var result int
	err := m.DB.QueryRow("SELECT uid FROM `forget_passw` WHERE uri = ? AND superseded = 0", uri).Scan(&result)
	if err != nil {
		return 0, err
	}

	return result, nil
}

func (m *UserModel) NewPassword(password string, id int) error {
	_, err := m.DB.Exec("UPDATE `users` SET password = PASSWORD(?) WHERE id = ?", password, id)
	if err != nil {
		return err
	}

	_, _ = m.DB.Exec("UPDATE `forget_passw` SET superseded =1 WHERE uid = ?", id)
	return nil
}

func (m *UserModel) CheckBearerToken(token string) string {
	if token == "" {
		return ""
	}

	return strings.TrimPrefix(token, "Bearer ")
}

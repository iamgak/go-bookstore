package models

import (
	"database/sql"
	// "log"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Id    int
	Email string
}

// to use main db that initialised in main.go
type UserModel struct {
	DB *sql.DB
}

func (m *UserModel) InsertUser(email, password, hashed string) (bool, error) {
	HashedPassword, err := m.GeneratePassword(password)
	if err != nil {
		return false, err
	}

	_, err = m.DB.Exec("INSERT INTO users(`email`,`password`,`activation_token`) VALUES (?, ?,? )", email, string(HashedPassword), hashed)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (m *UserModel) CreateAuthHeader(token string, uid int) error {
	_, err := m.DB.Exec("UPDATE `users` SET `login_token` = ? WHERE `id` = ?", token, uid)
	if err != nil {
		return err
	}
	return nil
}

func (m *UserModel) Logout(token string) error {
	_, err := m.DB.Exec("UPDATE `users` SET `login_token` = NULL WHERE `token` = ?", token)
	if err != nil {
		return err
	}
	return nil
}
func (m *UserModel) Authenticate(email, password string) (int, error) {
	var databasePassword string
	var uid int
	err := m.DB.QueryRow("SELECT password, id FROM `users` WHERE `active` = 1 AND `email` = ? ", strings.TrimSpace(email)).Scan(&databasePassword, &uid)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(databasePassword), []byte(password))
	// err = bcrypt.CompareHashAndPassword(currentHashedPassword, []byte(currentPassword))
	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return 0, nil
		}
		return 0, err

	}

	return uid, err
}

func (m *UserModel) EmailExist(email string) int {
	var uid int
	_ = m.DB.QueryRow("SELECT `id` FROM `users` WHERE  `email` = ?", email).Scan(&uid)
	return uid
}

func (m *UserModel) ValidToken(token string) int {
	var id int
	_ = m.DB.QueryRow("SELECT `id` FROM `users` WHERE `login_token` = ? ", token).Scan(&id)
	return id
}

func (m *UserModel) ValidURI(uri string) bool {
	var exists int
	query := "SELECT 1 FROM users WHERE activation_token = ? AND active = 0"
	err := m.DB.QueryRow(query, uri).Scan(&exists)
	if err != nil {
		return false
	}

	return exists > 0
}

func (m *UserModel) AccountActivate(token string) error {
	_, err := m.DB.Exec("UPDATE `users` SET `activation_token` = NULL, `active` = 1 WHERE `activation_token` = ? ", token)
	return err
}

func (m *UserModel) ForgetPassword(uid int, uri string) error {
	_, _ = m.DB.Exec("UPDATE `forget_passw` SET `superseded` = 1 WHERE `uid` = ?", uid)
	_, err := m.DB.Exec("INSERT INTO `forget_passw` (`uid`,`uri`,`superseded`) VALUES(?,?,0) ", uid, uri)
	return err
}

func (m *UserModel) ForgetPasswordUri(uri string) (int, error) {
	var result int
	err := m.DB.QueryRow("SELECT uid FROM `forget_passw` WHERE `uri` = ? AND `superseded` = 0", uri).Scan(&result)
	if err != nil {
		return 0, err
	}

	return result, nil
}

func (m *UserModel) NewPassword(newPassword string, id int) error {
	newHashedPassword, err := m.GeneratePassword(newPassword)
	if err != nil {
		return err
	}

	stmt := "UPDATE users SET password = ? WHERE id = ?"
	_, err = m.DB.Exec(stmt, string(newHashedPassword), id)
	if err != nil {
		return err
	}

	_, _ = m.DB.Exec("UPDATE `forget_passw` SET `superseded` =1 WHERE `uid` = ?", id)
	return nil
}

func (m *UserModel) GeneratePassword(newPassword string) ([]byte, error) {
	newHashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), 12)
	return newHashedPassword, err
}
func (m *UserModel) CheckBearerToken(token string) string {
	if token == "" {
		return ""
	}

	return strings.TrimPrefix(token, "Bearer ")
}

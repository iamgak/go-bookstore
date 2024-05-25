package models

import (
	"database/sql"
	"errors"
)

type Book struct {
	ISBN         string  `json:"isbn"`
	Title        string  `json:"title"`
	Author       string  `json:"author"`
	Price        float32 `json:"price"`
	Descriptions string  `json:"descriptions"`
	Genre        string  `json:"genre"`
}

type BookModel struct {
	DB *sql.DB
}

func (m *BookModel) InsertBook(ISBN, Author, Title, Genre, Descriptions string, Price float64, User_id int) (bool, error) {
	result, err := m.DB.Exec("INSERT INTO `reviews` (`isbn`,`price`,`title`,`author`,`genre`,`descriptions`,`uid`) VALUES (?,?,?,?,?,?,? )", ISBN, Price, Title, Author, Genre, Descriptions, User_id)
	if err != nil {
		return false, err
	}

	_, err = result.LastInsertId()
	if err != nil {
		return false, err
	}

	return false, nil
}

func (m *BookModel) BookExist(ISBN string) (int64, error) {
	var valid int64
	err := m.DB.QueryRow("SELECT 1 FROM `reviews` WHERE  `ISBN` = ?", ISBN).Scan(&valid)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		} else {
			return 0, err

		}
	}

	if valid == 0 {
		return 0, nil
	}

	return valid, nil
}

func (m *BookModel) GET(ISBN string) (*Book, error) {
	bk := &Book{}
	err := m.DB.QueryRow("SELECT isbn, title, author, price, descriptions, genre FROM `reviews` WHERE  `ISBN` = ?", ISBN).Scan(&bk.ISBN, &bk.Title, &bk.Author, &bk.Price, &bk.Descriptions, &bk.Genre)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		} else {
			return nil, err
		}
	}

	return bk, nil
}

func (m *BookModel) BooksListing() ([]*Book, error) {
	rows, err := m.DB.Query("SELECT isbn, title, author, price, descriptions,genre FROM `reviews`")
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		} else {
			return nil, err
		}
	}

	defer rows.Close()
	bks := []*Book{}
	for rows.Next() {
		bk := &Book{}
		_ = rows.Scan(&bk.ISBN, &bk.Title, &bk.Author, &bk.Price, &bk.Descriptions, &bk.Genre)
		bks = append(bks, bk)
	}

	if err = rows.Err(); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		} else {
			return nil, err
		}
	}

	return bks, err
}

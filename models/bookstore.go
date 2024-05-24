package models

import (
	"database/sql"
	"errors"
)

type Book struct {
	ISBN   string  `json:"isbn"`
	Title  string  `json:"title"`
	Author string  `json:"author"`
	Price  float32 `json:"price"`
}

// Define a new BookModel type which wraps a database connection pool.
type BookModel struct {
	DB *sql.DB
}

// We'll use the Insert method to add a new record to the "bookstore" table.
func (m *BookModel) InsertBook(ISBN, Author, Title string, Price float64) (int, error) {

	result, err := m.DB.Exec("INSERT INTO `books` (`ISBN`,`price`,`Title`,`author`) VALUES (?,?,?,? )", ISBN, Price, Title, Author)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

func (m *BookModel) BookExist(ISBN string) (int64, error) {
	var valid int64
	err := m.DB.QueryRow("SELECT 1 FROM `books` WHERE  `ISBN` = ?", ISBN).Scan(&valid)
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
	err := m.DB.QueryRow("SELECT ISBN,Title,author,price FROM `books` WHERE  `ISBN` = ?", ISBN).Scan(&bk.ISBN, &bk.Title, &bk.Author, &bk.Price)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		} else {
			return nil, err
		}
	}

	return bk, err

}

func (m *BookModel) BooksListing() ([]*Book, error) {

	rows, err := m.DB.Query("SELECT isbn, title, author, price FROM `books`")
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecord
		} else {
			return nil, err
		}
	}
	defer rows.Close()

	bks := []*Book{}
	for rows.Next() {
		bk := &Book{}
		_ = rows.Scan(&bk.ISBN, &bk.Title, &bk.Author, &bk.Price)

		bks = append(bks, bk)
	}

	if err = rows.Err(); err != nil {

		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecord
		} else {
			return nil, err

		}

	}
	return bks, err
}
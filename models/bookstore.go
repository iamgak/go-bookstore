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

// add books in db
func (m *BookModel) CreateBook(book *Book) error {
	_, err := m.DB.Exec("INSERT INTO `books` (`isbn`,`title`,`author`,`price`,`descriptions`,`genre`) VALUES (?,?,?,?,?,? )", &book.ISBN, &book.Title, &book.Author, &book.Price, &book.Descriptions, &book.Genre)
	return err
}

// check isbn already exist or not()
func (m *BookModel) BookExist(ISBN string) bool {
	var valid int
	_ = m.DB.QueryRow("SELECT 1 FROM `books` WHERE  `isbn` = ?", ISBN).Scan(&valid)
	return valid > 0
}

func (m *BookModel) GET(ISBN string) (*Book, error) {
	bk := &Book{}
	err := m.DB.QueryRow("SELECT isbn, title, author, price, descriptions, genre FROM `books` WHERE  `ISBN` = ?", ISBN).Scan(&bk.ISBN, &bk.Title, &bk.Author, &bk.Price, &bk.Descriptions, &bk.Genre)
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
	stmt := "SELECT isbn, title, genre, price, descriptions FROM books"
	Books, err := m.Listing(stmt)
	return Books, err
}

func (m *BookModel) Listing(stmt string) ([]*Book, error) {
	rows, err := m.DB.Query(stmt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		} else {
			return nil, err
		}
	}

	defer rows.Close()

	Books := []*Book{}
	for rows.Next() {
		bk, err := m.ScanBookData(rows)
		if err != nil {
			return nil, err
		}

		Books = append(Books, bk)
	}

	return Books, err
}

func (m *BookModel) GetBookByIsbn(isbn string) ([]*Book, error) {
	stmt := "SELECT isbn, title, author, genre, price, descriptions FROM `Books`"
	Book, err := m.Listing(stmt)
	return Book, err
}

func (m *BookModel) ScanBookData(rows *sql.Rows) (*Book, error) {
	book := new(Book)
	err := rows.Scan(
		&book.ISBN,
		&book.Title,
		&book.Author,
		&book.Genre,
		&book.Price,
		&book.Descriptions)
	return book, err
}

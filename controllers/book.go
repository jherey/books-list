package controllers

import (
	"encoding/json"
	"log"
	"net/http"
	"database/sql"
	"github.com/gorilla/mux"
	"github.com/jherey/books-list/models"
)

type Controller struct {}

var books []models.Book

func logFatal(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func (c Controller) GetBooks(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Get all books")
		var book models.Book
		books = []models.Book{}

		rows, err := db.Query("select * from books")
		logFatal(err)

		defer rows.Close()

		for rows.Next() {
			err := rows.Scan(&book.ID, &book.Title, &book.Author, &book.Year)
			logFatal(err)

			books = append(books, book)
		}

		json.NewEncoder(w).Encode(books)
	}
}

func (c Controller) GetBook(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Get one book")
		var book models.Book
		params := mux.Vars(r)

		rows := db.QueryRow("select * from books where id=$1", params["id"])

		err := rows.Scan(&book.ID, &book.Title, &book.Author, &book.Year)
		logFatal(err)

		json.NewEncoder(w).Encode(book)
	}
}

func (c Controller) AddBook(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Adds one book")
		var book models.Book
		var bookID int

		json.NewDecoder(r.Body).Decode(&book)

		err := db.QueryRow("insert into books (title, author, year) values($1, $2, $3) RETURNING ID;",
			book.Title, book.Author, book.Year).Scan(&bookID)

		logFatal(err)

		json.NewEncoder(w).Encode(bookID)
	}
}

func (c Controller) UpdateBook(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Updates a book")

		var book models.Book
		json.NewDecoder(r.Body).Decode(&book)

		result, err := db.Exec("update books set title=$1, author=$2, year=$3 where id=$4 RETURNING id",
			&book.Title, &book.Author, &book.Year, &book.ID)

		rowsUpdated, err := result.RowsAffected()
		logFatal(err)

		json.NewEncoder(w).Encode(rowsUpdated)
	}
}

func (c Controller) RemoveBook(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Removes a book")
		params := mux.Vars(r)

		result, err := db.Exec("delete from books where id=$1", params["id"])
		logFatal(err)

		rowsDeleted, err := result.RowsAffected()
		logFatal(err)

		json.NewEncoder(w).Encode(rowsDeleted)
	}
}

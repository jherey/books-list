package main

import (
	"os"
	"encoding/json"
	"strconv"
	"log"
	"net/http"
	"database/sql"
	"github.com/gorilla/mux"
	"github.com/lib/pq"
	"github.com/subosito/gotenv"
)

type Book struct {
	ID 			int				`json:id`
	Title 	string		`json:title`
	Author	string		`json:author`
	Year 		string		`json:year`
}

var books []Book
var db *sql.DB

func init() {
	gotenv.Load()
}

func logFatal(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	pgURL, err := pq.ParseURL(os.Getenv("ELEPHANTSQL_URL"))
	logFatal(err)

	db, err = sql.Open("postgres", pgURL)
	logFatal(err)

	err = db.Ping()
	logFatal(err)

	log.Println(pgURL)

	router := mux.NewRouter()

	router.HandleFunc("/books", getBooks).Methods("GET")
	router.HandleFunc("/books/{id}", getBook).Methods("GET")
	router.HandleFunc("/books", addBook).Methods("POST")
	router.HandleFunc("/books", updateBook).Methods("PUT")
	router.HandleFunc("/books/{id}", removeBook).Methods("DELETE")

	log.Fatal(http.ListenAndServe(":8000", router)) // log errors
}

func getBooks(w http.ResponseWriter, r *http.Request) {
	log.Println("Get all books")
	var book Book
	books = []Book{}

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

func getBook(w http.ResponseWriter, r *http.Request) {
	log.Println("Get one book")
	var book Book
	params := mux.Vars(r)

	rows := db.QueryRow("select * from books where id=$1", params["id"])

	err := rows.Scan(&book.ID, &book.Title, &book.Author, &book.Year)
	logFatal(err)

	json.NewEncoder(w).Encode(book)
}

func addBook(w http.ResponseWriter, r *http.Request) {
	log.Println("Adds one book")
	var book Book
	var bookID int

	json.NewDecoder(r.Body).Decode(&book)

	err := db.QueryRow("insert into books (title, author, year) values($1, $2, $3) RETURNING ID;",
		book.Title, book.Author, book.Year).Scan(&bookID)

	logFatal(err)

	json.NewEncoder(w).Encode(bookID)
}

func updateBook(w http.ResponseWriter, r *http.Request) {
	log.Println("Updates a book")

	var book Book
	json.NewDecoder(r.Body).Decode(&book)

	result, err := db.Exec("update books set title=$1, author=$2, year=$3 where id=$4 RETURNING id",
		&book.Title, &book.Author, &book.Year, &book.ID)

	rowsUpdated, err := result.RowsAffected()
	logFatal(err)

	json.NewEncoder(w).Encode(rowsUpdated)
}

func removeBook(w http.ResponseWriter, r *http.Request) {
	log.Println("Removes a book")
	params := mux.Vars(r)
	id, _ := strconv.Atoi(params["id"])

	for i, item := range books {
		if item.ID == id {
			books = append(books[:i], books[i+1:]...)
		}
	}

	json.NewEncoder(w).Encode(books)
}

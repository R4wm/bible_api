package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"strconv"

	"net/http"

	"github.com/r4wm/kjvapi"
)

var DB *sql.DB

type response struct {
	Text string `json:"text"`
}

// helloWorld basic handler function
func helloWorld(w http.ResponseWriter, r *http.Request) {
	basicResponse := &response{Text: "Hello World " + r.URL.Path[:]}
	jsonResponse, _ := json.Marshal(basicResponse)

	w.Header().Set("Content-Type", "application/json")

	fmt.Fprintf(w, string(jsonResponse))
}

func gimmeVerse(w http.ResponseWriter, r *http.Request) {

	//Static stuff for now
	verseResponse := kjvapi.KJVChapter{}
	verseResponse.Chapter = 1

	rows, _ := DB.Query("select verse, text from kjv where book = 'Genesis' and chapter = 1")

	var text string
	var verse int

	for rows.Next() {
		rows.Scan(&verse, &text)
		verseResponse.Verses = append(verseResponse.Verses,
			kjvapi.KJVVerse{
				Verse: verse,
				Text:  text})

	}

	jsonResponse, err := json.Marshal(verseResponse)

	// TODO Change this to a 500 response or something
	if err != nil {
		panic(err)
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, string(jsonResponse))
}

// GetBooks retrieve list of books from the kjv db
func GetBooks(w http.ResponseWriter, r *http.Request) {
	var books []string
	var bookName string

	rows, _ := DB.Query("select distinct book from kjv")

	for rows.Next() {
		rows.Scan(&bookName)
		books = append(books, bookName)
	}

	jsonResponse, err := json.Marshal(books)

	if err != nil {
		log.Fatal("Could not marshal books")
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, string(jsonResponse))
}

// GetChapter print the book, chapter and verses in json format
func GetChapter(w http.ResponseWriter, r *http.Request) {
	var verses []kjvapi.KJVVerse

	book, ok := r.URL.Query()["book"]
	if !ok || len(book[0]) < 1 {
		log.Println("Url param book is missing.")
		w.WriteHeader(http.StatusNotAcceptable)
		w.Write([]byte("406 - book param not found."))
		return
	}

	chapter, ok := r.URL.Query()["chapter"]
	if !ok {
		log.Println("Url param chapter is missing.")
		w.WriteHeader(http.StatusNotAcceptable)
		w.Write([]byte("406 - chapter param not found."))
		return
	}

	stmt := fmt.Sprintf("select verse, text from kjv where book='%s' and chapter=%v", book[0], chapter[0])

	rows, err := DB.Query(stmt)
	if err != nil {
		log.Println(err)
		log.Printf("database: %#v\n", DB)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("400 - Could not query such a request: "))
		return
	}

	var verse int
	var text string

	for rows.Next() {
		rows.Scan(&verse, &text)
		verses = append(verses, kjvapi.KJVVerse{Verse: verse, Text: text})
	}
	i, err := strconv.Atoi(chapter[0])
	if err != nil {
		log.Printf("Could not convert %s to int.", chapter[0])
	}

	bkResult := kjvapi.KJVBook{
		Book: book[0],
		Chapters: []kjvapi.KJVChapter{
			kjvapi.KJVChapter{Chapter: i, Verses: verses}}}

	response, _ := json.Marshal(bkResult)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, string(response))
}

func main() {
	/////////////////
	// Args	       //
	/////////////////
	createDB := flag.Bool("createDB", false, "create database")
	dbPath := flag.String("dbPath", "", "path to datebase")
	flag.Parse()
	/////////////////////
	// CreateDB	   //
	/////////////////////
	if *createDB && len(*dbPath) > 0 {
		kjvapi.CreateKJVDB(*dbPath)

	}
	////////////////////////////////
	// Database Connection	      //
	////////////////////////////////
	DB, _ = sql.Open("sqlite3", *dbPath)
	fmt.Println(fmt.Sprintf("%T\n", DB))
	/////////////////////
	// Handlers	   //
	/////////////////////
	http.HandleFunc("/", helloWorld)
	http.HandleFunc("/gimmeVerse", gimmeVerse)
	http.HandleFunc("/get_books", GetBooks)
	http.HandleFunc("/get_chapter", GetChapter)
	http.ListenAndServe(":8000", nil)
}

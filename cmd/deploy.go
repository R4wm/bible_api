package main

/////////////
// Imports //
/////////////
import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/r4wm/kjvapi"
	"github.com/r4wm/mintz5"
)

/////////////
// Structs //
/////////////
//Book Name of Book and how many chapters contained in that book.
type Book struct {
	Name     string
	Chapters int
}

// KJVMapping static mapping containing books and number of chapters per book.
type KJVMapping struct {
	Books []Book
}

type response struct {
	Text string `json:"text"`
}

// RandVerse struct for get_random_verse endpoint.
type RandVerse struct {
	Book    string `json:"Book"`
	Chapter int    `json:"Chapter"`
	Verse   int    `json:"Verse"`
	Text    string `json:"Text"`
}

//////////
// Vars //
//////////
var DB *sql.DB
var Mapping KJVMapping

///////////////
// Functions //
///////////////
// GetBooks retrieve list of books from the kjv db
func GetBooks(w http.ResponseWriter, r *http.Request) {
	jsonResponse, err := json.Marshal(Mapping)

	if err != nil {
		log.Fatal("Could not marshal books")
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, string(jsonResponse))
}

// GetChapter print the book, chapter and verses in json format
func GetChapter(w http.ResponseWriter, r *http.Request) {
	log.Printf("%#v\n", r)

	var verses []kjvapi.KJVVerse

	book, ok := r.URL.Query()["book"]
	if !ok || len(book[0]) < 1 {
		log.Println("Url param book is missing.")
		w.WriteHeader(http.StatusNotAcceptable)
		w.Write([]byte("406 - book param not found."))
		return
	}

	book[0] = strings.ToUpper(book[0])

	chapter, ok := r.URL.Query()["chapter"]
	if !ok {
		log.Println("Url param chapter is missing.")
		w.WriteHeader(http.StatusNotAcceptable)
		w.Write([]byte("406 - chapter param not found."))
		return
	}

	stmt := fmt.Sprintf("select verse, text from kjv where book='%s' and chapter=%v", book[0], chapter[0])

	rows, err := DB.Query(stmt)
	defer rows.Close()

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

func GetVerse(w http.ResponseWriter, r *http.Request) {

	var verse kjvapi.KJVVerse

	neededItems := []string{"book", "chapter", "verse"}
	for _, item := range neededItems {
		_, ok := r.URL.Query()[item]
		if !ok {
			w.WriteHeader(http.StatusNotAcceptable)
			msg := fmt.Sprintf("%s arg required.\n", item)
			w.Write([]byte(msg))
			return
		}

		//Args check out ok.
		stmt := fmt.Sprintf("select text from kjv where book=%s and chapter=%s and verse=%s",
			strings.ToUpper(r.URL.Query()["book"][0]),
			r.URL.Query()["chapter"][0],
			r.URL.Query()["verse"][0])

		rows, err := DB.Query(stmt)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Header().Set("Content-Type", "application/text")
			w.Write([]byte("Could not query database."))
			return
		}
		defer rows.Close()

		for rows.Next() {
			rows.Scan(&verse.Text)
		}

		if len(verse.Text) <= 0 {
			log.Printf("Got nothing from database: %s\n", stmt)
		} else {
			verse.Verse, err = strconv.Atoi(r.URL.Query()["verse"][0])
			if err != nil {
				log.Printf("Could NOT convert %v to int\n",
					r.URL.Query()["verse"][0])
			}
		}
	}

	// Return the result
	result, _ := json.Marshal(verse)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(result))
	fmt.Println(verse)
}

func GetRandomVerseFromDB() (RandVerse, error) {
	const lastCardinalVerseNum = 31101

	var randVerse RandVerse

	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)

	stmt := fmt.Sprintf("select book, chapter, verse, text from kjv where ordinal_verse=%d",
		r1.Intn(lastCardinalVerseNum))

	rows, err := DB.Query(stmt)
	if err != nil {
		log.Fatalf("Failed DB.Query(%s)\n", stmt)
		return randVerse, err
	}

	for rows.Next() {
		rows.Scan(&randVerse.Book, &randVerse.Chapter, &randVerse.Verse, &randVerse.Text)
	}

	// OK
	return randVerse, nil

	// fmt.Printf("Type randVerse: %T\n", randVerse)

	// result, err := json.Marshal(randVerse)
	// if err != nil {
	// 	log.Printf("Could not json marshal %#v\n", randVerse)
	// 	return result, err
	// }

	// println("OK")

	// return result, err
}

// GetRandomVerseAPI returns json formatted true random verse.
func GetRandomVerseAPI(w http.ResponseWriter, r *http.Request) {

	// get the verse struct
	rv, err := GetRandomVerseFromDB()
	if err != nil {
		w.Header().Set("Content-Type", "application/text")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to get random verse from db."))
		return
	}

	// marshal to json
	randomVerse, err := json.Marshal(rv)
	if err != nil {
		w.Header().Set("Content-Type", "application/text")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to Marshal data to json."))
		return
	}

	// OK serve it
	log.Printf("Endpoint: %s IP: %s -> %s\n", r.URL, r.RemoteAddr, randomVerse)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(randomVerse))
}

// GetRandomVerse write pretty html page with random verse.
func GetRandomVerse(w http.ResponseWriter, r *http.Request) {

	result, err := GetRandomVerseFromDB()
	if err != nil {

		w.Header().Set("Content-Type", "application/text")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to get random verse from db."))
		return
	}

	// create the data struct for template
	returnPage := struct {
		Verse RandVerse
		Color string
	}{
		result,
		mintz5.GetRandomColor(),
	}

	// logging
	log.Printf("Endpoint: %s IP: %s -> %s\n", r.URL, r.RemoteAddr, result)

	// TODO Move this to file and cache read 1 time and reuse..
	tmpl, err := template.New("Basic").Parse(
		`
<!DOCTYPE html>
<html>
<body style="background-color:{{ .Color }};">
<h1><center>{{ .Verse.Book }} {{ .Verse.Chapter }}:{{ .Verse.Verse }} </center></h1>
<h3><center>{{ .Verse.Text }}</center></h3>
</body>
</html>`)

	if err != nil {
		w.Header().Set("Content-Type", "application/text")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to parse template"))
		log.Println(err)
		return
	}

	// Ok Serve it.
	err = tmpl.Execute(w, returnPage)
	if err != nil {
		w.Header().Set("Content-Type", "application/text")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to write to Writer"))
		log.Println(err)
		return
	}
}

func main() {
	fmt.Printf("Mapping: %#v\n", Mapping)

	/////////////////
	// Args	       //
	/////////////////
	createDB := flag.Bool("createDB", false, "create database")
	dbPath := flag.String("dbPath", "", "path to datebase")
	flag.Parse()

	if len(*dbPath) == 0 {
		log.Fatalf("Must provide dbPath")
	}

	if _, err := os.Stat(*dbPath); os.IsNotExist(err) {
		if *createDB == false {
			log.Fatalf("database file does not exist: %s\n", *dbPath)
		} else {
			kjvapi.CreateKJVDB(*dbPath)
		}
	}

	fmt.Println("dbPath: ", *dbPath)
	fmt.Println("createDB: ", *createDB)
	////////////////////////////////
	// Database Connection	      //
	////////////////////////////////
	DB, _ = sql.Open("sqlite3", *dbPath)
	fmt.Println(fmt.Sprintf("%T\n", DB))
	log.Printf("Running server using database at: %s\n", *dbPath)

	/////////////////////////////
	// Populate Mapping	   //
	/////////////////////////////
	// Cant do this part in an init() cause it will run before main and we havent spec'd the db from args
	// TODO: Maybe make db location fixed..
	// populate the Book struct
	rows, _ := DB.Query("select distinct book from kjv")
	defer rows.Close()

	for rows.Next() {
		var bookName string
		rows.Scan(&bookName)
		book := Book{Name: bookName}

		chaptersQuery := fmt.Sprintf("select max(chapter) from kjv where book=\"%s\"", bookName)
		fmt.Println(chaptersQuery)
		rowsForChapterCount, err := DB.Query(chaptersQuery)
		defer rowsForChapterCount.Close()

		if err != nil {
			log.Fatalf("Failed query on %s\n", chaptersQuery)
		}

		for rowsForChapterCount.Next() {
			err := rowsForChapterCount.Scan(&book.Chapters)
			if err != nil {
				log.Fatalf("Could not get %s from db.\n", bookName)
			}

			Mapping.Books = append(Mapping.Books, book)
		}

	}
	fmt.Printf("Mapping: %#v\n", Mapping)

	////////////////////////////////////////////////////////////////////
	// HANDLERS							  //
	// NOTE: the "/api/*" are raw api json formatted endpoints	  //
	////////////////////////////////////////////////////////////////////
	http.HandleFunc("/get_books", GetBooks)
	http.HandleFunc("/get_chapter", GetChapter)
	http.HandleFunc("/get_verse", GetVerse)
	http.HandleFunc("/get_random_verse", GetRandomVerse)
	http.HandleFunc("/api/get_random_verse", GetRandomVerseAPI)
	http.ListenAndServe(":8000", nil)
}

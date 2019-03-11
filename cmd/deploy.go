package main

// TODO:
// replace manual http.Response stuff with http.Error
// find some way to alias long things like r.URL.Query()[intable][0]
// uniform the endpoints, json, webpage etc.. its all mixed up.

/////////////
// Imports //
/////////////
import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/r4wm/kjvapi"
	"github.com/r4wm/mintz5"
	log "github.com/sirupsen/logrus"
)

const (
	hostname = "http://cdn.mintz5.com/801A6BD/linode"
	// hostname = "http://mintz5.com:8000"
	// hostname = `http://localhost:8000`

	chapterTemplate = `
<html>
  <body style="background-color:{{ .Color }};">
    <h1><center>{{ .BookName }} {{ .Chapter }}</h1>
  <body>
    {{ range $index, $results := .Verses }}
    <p><b><left><a href={{ verseLink $index }}> {{ add $index 1}}</a> {{ . }} </b></p>
    {{ end }}
  </body>
</html> 
`

	verseTemplate = `
<!DOCTYPE html>
<html>
   <body style="background-color:{{ .Color }};">
      <h1>
	 <center>{{ .Verse.Book }} {{ .Verse.Chapter }}:{{ .Verse.Verse }} </center>
      </h1>
      <h3>
	 <center>{{ .Verse.Text }}</center>
      </h3>
   </body>
</html>
`

	chapterButtonsTemplate = `
<!DOCTYPE html>
<html>
   <head>
      <meta name="viewport" content="width=device-width, initial-scale=1">
      <style>
	 .block {
	 display: block;
	 width: 100%;
	 border: none;
	 background-color: #4CAF50;
	 color: white;
	 padding: 14px 28px;
	 font-size: 16px;
	 cursor: pointer;
	 text-align: center;
	 }
	 .block:hover {
	 background-color: #ddd;
	 color: black;
	 }
      </style>
      <title>{{ .Name }}</title>
   </head>
   <body style="background-color:{{ .Color }};">
     <p><center><h1> {{ .Name }} </h1><center></p>
     {{ range $index, $results := .Links }}
       <p><button class="block" onclick="window.location.href={{ $results }};">{{ add $index 1 }}</button></p>
     {{ end }}
   </body>
</html>
`

	booksButtonsTemplate = `
<!DOCTYPE html>
<html>
   <head>
      <meta name="viewport" content="width=device-width, initial-scale=1">
      <style>
	 .block {
	 display: block;
	 width: 100%;
	 border: none;
	 background-color: #4CAF50;
	 color: white;
	 padding: 14px 28px;
	 font-size: 16px;
	 cursor: pointer;
	 text-align: center;
	 }
	 .block:hover {
	 background-color: #ddd;
	 color: black;
	 }
      </style>
      <title>Books of the Bible</title>
   </head>
   <body style="background-color:{{ .Color }};">
      {{ range $key, $value := .Books }}
      <p><button class="block" onclick="window.location.href={{ createLink $value }};">{{ $value }}</button></p>
      {{ end }}
   </body>
</html>
`
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

// ListBooks uses booksButtonstemplate to layout the books in html.
func ListBooks(w http.ResponseWriter, r *http.Request) {

	log.Info("ListBooks")

	books := []string{
		"GENESIS",
		"EXODUS",
		"LEVITICUS",
		"NUMBERS",
		"DEUTERONOMY",
		"JOSHUA",
		"JUDGES",
		"RUTH",
		"1SAMUEL",
		"2SAMUEL",
		"1KINGS",
		"2KINGS",
		"1CHRONICLES",
		"2CHRONICLES",
		"EZRA",
		"NEHEMIAH",
		"ESTHER",
		"JOB",
		"PSALMS",
		"PROVERBS",
		"ECCLESIASTES",
		"SONG OF SOLOMON",
		"ISAIAH",
		"JEREMIAH",
		"LAMENTATIONS",
		"EZEKIEL",
		"DANIEL",
		"HOSEA",
		"JOEL",
		"AMOS",
		"OBADIAH",
		"JONAH",
		"MICAH",
		"NAHUM",
		"HABAKKUK",
		"ZEPHANIAH",
		"HAGGAI",
		"ZECHARIAH",
		"MALACHI",
		"MATTHEW",
		"MARK",
		"LUKE",
		"JOHN",
		"ACTS",
		"ROMANS",
		"1CORINTHIANS",
		"2CORINTHIANS",
		"GALATIANS",
		"EPHESIANS",
		"PHILIPPIANS",
		"COLOSSIANS",
		"1THESSALONIANS",
		"2THESSALONIANS",
		"1TIMOTHY",
		"2TIMOTHY",
		"TITUS",
		"PHILEMON",
		"HEBREWS",
		"JAMES",
		"1PETER",
		"2PETER",
		"1JOHN",
		"2JOHN",
		"3JOHN",
		"JUDE",
		"REVELATION"}

	// funcs generates the link needed for button
	funcs := template.FuncMap{"createLink": func(b string) string {
		return fmt.Sprintf("%s/list_chapters?book=%s", hostname, b)
	}}

	t, err := template.New("listBooks").Funcs(funcs).Parse(booksButtonsTemplate)
	if err != nil {
		log.Warnf("Could not list books: %s\n", err)
	}

	booksStruct := struct {
		Books []string
		Color string
	}{
		Books: books,
		Color: mintz5.GetRandomColor(),
	}

	t.Execute(w, booksStruct)
}

// ListChapters list the chapters of the book with clickable buttons for navigation
func ListChapters(w http.ResponseWriter, r *http.Request) {
	log.Infof("ListChapters %s\n", r.RequestURI)
	log.Infof("ListChapters: %s\n", r)

	book, ok := r.URL.Query()["book"]

	// Handle missing book args
	if !ok || len(book) < 1 {
		log.Warnf("Url param book is missing: %v\n", book)
		w.WriteHeader(http.StatusNotAcceptable)
		w.Write([]byte("406 - book param not found."))
		return
	}

	chaptersMax := mintz5.BookChapterLimit[strings.ToUpper(book[0])]
	// TODO How should this be handled??
	if chaptersMax == 0 {
		log.Warnf("Book has no chapters. possible a typo if manually entered: %s\n")
	}

	chapterInfo := struct {
		Name     string
		Chapters []int
		Links    []string
		Color    string
	}{
		Name:  strings.ToUpper(book[0]),
		Color: mintz5.GetRandomColor(),
	}

	// populate the data by interation
	for i := 1; i <= chaptersMax; i++ {
		chapterInfo.Chapters = append(chapterInfo.Chapters, i)
		chapterInfo.Links = append(
			chapterInfo.Links,
			fmt.Sprintf("%s/get_chapter?book=%s&chapter=%s", hostname, chapterInfo.Name, strconv.Itoa(i)))
	}

	funcs := template.FuncMap{"add": func(x, y int) int { return x + y }}
	t, err := template.New("chapterListing").Funcs(funcs).Parse(chapterButtonsTemplate)
	if err != nil {
		log.Errorf("ListBooks failed : %v\n", err)
	}

	w.Header().Set("Content-Type", "text/html")
	t.Execute(w, chapterInfo)
}

// GetBooks retrieve list of books from the kjv db
func GetBooks(w http.ResponseWriter, r *http.Request) {
	jsonResponse, err := json.Marshal(Mapping)

	if err != nil {
		log.Fatal("Could not marshal books")
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, string(jsonResponse))
}

// GetChapterAPI print the book, chapter and verses in json format
func GetChapterAPI(w http.ResponseWriter, r *http.Request) {
	log.Infof("%#v\n", r)

	verses := struct {
		BookName string
		Chapter  int
		Verses   []string
		Color    string
	}{Color: mintz5.GetRandomColor()}

	// var verses []kjvapi.KJVVerse

	// Check the Book arg from request.
	book, ok := r.URL.Query()["book"]
	if !ok || len(book[0]) < 1 {
		log.Println("Url param book is missing.")
		w.WriteHeader(http.StatusNotAcceptable)
		w.Write([]byte("406 - book param not found."))
		return
	}

	book[0] = strings.ToUpper(book[0])
	verses.BookName = book[0]

	// Check the chapter arg in request
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
		verses.Verses = append(verses.Verses, text)
	}

	i, err := strconv.Atoi(chapter[0])
	verses.Chapter = i
	if err != nil {
		log.Printf("Could not convert %s to int.", chapter[0])
	}

	//////////////////////////
	// Template time        //
	//////////////////////////
	// add function to increment range indexing since it starts at 0 by default
	funcs := template.FuncMap{"add": func(x, y int) int { return x + y }}
	verseLink := template.FuncMap{"verseLink": func(x int) string {

		verseOffSet := strconv.Itoa(x + 1)

		return fmt.Sprintf("%s/get_verse?book=%s&chapter=%s&verse=%s&json=false",
			hostname,
			book[0],
			chapter[0],
			verseOffSet,
		)
	}}

	t, err := template.New("chapter").Funcs(funcs).Funcs(verseLink).Parse(chapterTemplate)

	if err != nil {
		panic(err)
	}

	t.Execute(w, verses)
}

// GetVerse writes single verse to json response
func GetVerse(w http.ResponseWriter, r *http.Request) {

	var verse kjvapi.KJVVerse
	neededItems := []string{"book", "chapter", "verse", "json"}

	// check number of Args used
	if len(r.URL.Query()) != 4 {
		w.WriteHeader(http.StatusNotAcceptable)
		w.Write([]byte("Required args: book, chapter, verse"))
		return
	}

	// check expected args exist
	for _, item := range neededItems {
		_, ok := r.URL.Query()[item]

		if !ok {
			log.Warnf("Missing arg: %s\n", item)
			http.Error(w, fmt.Sprintf("%s arg required.\n", item), http.StatusBadRequest)
			return
		}
	}

	// check verse and chapter is "int"
	for _, intable := range []string{"chapter", "verse"} {
		if _, err := strconv.Atoi(r.URL.Query()[intable][0]); err != nil {
			msg := fmt.Sprintf("%s:%s is not an integer value\n", intable, r.URL.Query()[intable][0])
			log.Warnf(msg)
			http.Error(w, msg, http.StatusBadRequest)
			return
		}
	}

	//Args check out ok.
	log.Infof("book: %s\n", r.URL.Query()["book"][0])
	log.Infof("chapter: %s\n", r.URL.Query()["chapter"][0])
	log.Infof("verse: %s\n", r.URL.Query()["verse"][0])
	stmt := fmt.Sprintf("select text from kjv where book=\"%s\" and chapter=%s and verse=%s",
		strings.ToUpper(r.URL.Query()["book"][0]),
		r.URL.Query()["chapter"][0],
		r.URL.Query()["verse"][0])

	rows, err := DB.Query(stmt)

	// Handle no DB connect.
	if err != nil {
		http.Error(w, fmt.Sprintf("Could not query DB"), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Query from DB
	for rows.Next() {
		rows.Scan(&verse.Text)
	}

	// Check results from DB
	if len(verse.Text) <= 0 {
		log.Warnf("Got nothing from database: %s\n", stmt)

	} else {
		verse.Verse, err = strconv.Atoi(r.URL.Query()["verse"][0])
		if err != nil {
			log.Printf("Could NOT convert %v to int\n",
				r.URL.Query()["verse"][0])
		}
	}

	////////////////////////////
	// JSON OR WEBPAGE	  //
	////////////////////////////

	// Return json the result
	switch pretty := r.URL.Query()["json"][0]; pretty {
	case "true":
		result, _ := json.Marshal(verse)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(result))
		log.Info(verse)
		return

	// Return pretty web page
	case "false":
		var verse RandVerse

		// Prep DB for query
		stmt := fmt.Sprintf("select book, chapter, verse, text from kjv where book=\"%s\" and chapter=%s and verse=%s",
			strings.ToUpper(r.URL.Query()["book"][0]),
			r.URL.Query()["chapter"][0],
			r.URL.Query()["verse"][0])

		rows, err := DB.Query(stmt)

		if err != nil {
			log.Warnf("Failed DB.Query(%s)\n", stmt)
			http.Error(w, "Failed to Query DB", http.StatusInternalServerError)
			return
		}

		for rows.Next() {
			rows.Scan(&verse.Book, &verse.Chapter, &verse.Verse, &verse.Text)
		}

		// Make pretty web page from template
		tmpl, err := template.New("Basic").Parse(verseTemplate)

		// Handle template errors.
		if err != nil {
			w.Header().Set("Content-Type", "application/text")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Failed to parse template"))
			log.Println(err)
			return
		}

		// create the data struct for template
		returnPage := struct {
			Verse RandVerse
			Color string
		}{
			verse,
			mintz5.GetRandomColor(),
		}

		// Ok Serve it pretty
		err = tmpl.Execute(w, returnPage)
		if err != nil {
			w.Header().Set("Content-Type", "application/text")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Failed to write to Writer"))
			log.Println(err)
			return
		}

	// Handle bad json arg
	default:
		http.Error(w, fmt.Sprintf("json arg value not understood: %v\n", pretty), http.StatusInternalServerError)
		return
	}
}

// GetRandomVerseFromDB gets the verse from db to pass to pretty print api.
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

// GetVerseWebPage create nice web page from template
func GetVerseWebPage(w http.ResponseWriter, r *http.Request) {

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
	tmpl, err := template.New("Basic").Parse(verseTemplate)

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
	http.HandleFunc("/get_chapter", GetChapterAPI)
	http.HandleFunc("/get_verse", GetVerse)
	http.HandleFunc("/get_random_verse", GetRandomVerse)
	http.HandleFunc("/api/get_random_verse", GetRandomVerseAPI)
	http.HandleFunc("/list_chapters", ListChapters)
	http.HandleFunc("/list_books", ListBooks)
	http.ListenAndServe(":8000", nil)
}

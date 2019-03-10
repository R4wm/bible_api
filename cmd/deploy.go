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
    <p><b><left>{{ add $index 1}} {{ . }} </b></p>
    {{ end }}
      <p><button onclick="window.location.href=http://cdn.mintz5.com/801A6BD/linode/list_books;">BOOKS</button></p>
  </body>
</html>  `

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
      <title>Chapters for {{ .Name }} </title>
   </head>
   <body style="background-color:{{ .Color }};">
     <p><h1> {{ .Name }} </h1></p>
     {{ range $index, $results := .Links }}
       <p><button onclick="window.location.href={{ $results }};">{{ add $index 1 }}</button></p>
     {{ end }}
   </body>
</html>`

	booksButtonsTemplate = `
<!DOCTYPE html>
<html>
  <head>
    <title>Books of the Bible</title>
  </head>
  <body style="background-color:{{ .Color }};">
    {{ range $key, $value := .Books }}
    <p><button onclick="window.location.href={{ createLink $value }};">{{ $value }}</button></p>
    {{ end }}
  </body>
</html>`
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

	funcs := template.FuncMap{"createLink": func(b string) string { return fmt.Sprintf("%s/list_chapters?book=%s", hostname, b) }}

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

// ListBooks list the books of the Bible in button links
func ListChapters(w http.ResponseWriter, r *http.Request) {
	log.Info("ListChapters")
	book, ok := r.URL.Query()["book"]

	// Handle missing book args
	if !ok || len(book) < 1 {
		log.Warnf("Url param book is missing: %v\n", book)
		w.WriteHeader(http.StatusNotAcceptable)
		w.Write([]byte("406 - book param not found."))
		return
	}

	chaptersMax := mintz5.BookChapterLimit[strings.ToUpper(book[0])]
	log.Println(chaptersMax)

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

	t, err := template.New("chapter").Funcs(funcs).Parse(chapterTemplate)

	if err != nil {
		panic(err)
	}

	t.Execute(w, verses)
}

// GetVerse writes single verse to webpage template
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

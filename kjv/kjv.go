package kjv

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"text/template"
	"time"

	"net/http"

	"github.com/gorilla/mux"
	kjv "github.com/r4wm/mintz5"
)

const (
	lastCardinalVerseNum = 31101

	verseTemplate = `
<!DOCTYPE html>
<html>
   <body style="background-color:{{ .Color }};">
      <h1>
	 <center>
           <a href={{.ChapterRef}}>{{ .Verse.Book }} {{ .Verse.Chapter }}</a> : {{ .Verse.Verse }}
         </center>
      </h1>
      <h3>
	 <center>{{ .Verse.Text }}</center>
      </h3>
   </body>
</html>
`
	chapterTemplate = `
<html>
<style>
.btn-group button {
  background-color: gold; /* Green background */
  border: 1px solid green; /* Green border */
  color: black;
  padding: 10px 24px; /* Some padding */
  cursor: pointer; /* Pointer/hand icon */
  float: center; /* Float the buttons side by side */
}
/* Clear floats (clearfix hack) */
.btn-group:after {
  content: "";
  clear: both;
  display: table;
}
.btn-group button:not(:last-child) {
  border-right: none; /* Prevent double borders */
}
/* Add a background color on hover */
.btn-group button:hover {
  background-color: #3e8e41;
}
</style>
  <body style="background-color:{{ .Color }};">
    <h1><center>{{ .BookName }} {{ .Chapter }}</h1>
  <body>
    {{ range $index, $results := .Verses }}
    <p><b><left><a href={{ verseLink $index }}> {{ add $index 1}}</a> {{ . }} </b></p>
    {{ end }}
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link rel="stylesheet" href="https://www.w3schools.com/w3css/4/w3.css">
    <div class="w3-bar">
    <div class="btn-group">
    {{ if .PreviousChapterLink  }}
    <button onclick="window.location.href = '{{.PreviousChapterLink}}';" class="w3-bar-item w3-button" style="width:33.3%"> < </button>
    {{ end }}
    <button onclick="window.location.href = '{{.ListAllBooksLink}}';" class="w3-bar-item w3-button" style="width:33.3%">Books</button>
    {{ if .NextChapterLink  }}
    <button onclick="window.location.href = '{{.NextChapterLink}}';" class="w3-bar-item w3-button" style="width:33.3%"> > </button>
    {{ end }}
    </div>
  </body>
</html>
`
)

var (
	BookChapterLimit = map[string]int{
		"GENESIS":         50,
		"EXODUS":          40,
		"LEVITICUS":       27,
		"NUMBERS":         36,
		"DEUTERONOMY":     34,
		"JOSHUA":          24,
		"JUDGES":          21,
		"RUTH":            4,
		"1SAMUEL":         31,
		"2SAMUEL":         24,
		"1KINGS":          22,
		"2KINGS":          25,
		"1CHRONICLES":     29,
		"2CHRONICLES":     36,
		"EZRA":            10,
		"NEHEMIAH":        13,
		"ESTHER":          10,
		"JOB":             42,
		"PSALMS":          150,
		"PROVERBS":        31,
		"ECCLESIASTES":    12,
		"SONG OF SOLOMON": 8,
		"ISAIAH":          66,
		"JEREMIAH":        52,
		"LAMENTATIONS":    5,
		"EZEKIEL":         48,
		"DANIEL":          12,
		"HOSEA":           14,
		"JOEL":            3,
		"AMOS":            9,
		"OBADIAH":         1,
		"JONAH":           4,
		"MICAH":           7,
		"NAHUM":           3,
		"HABAKKUK":        3,
		"ZEPHANIAH":       3,
		"HAGGAI":          2,
		"ZECHARIAH":       14,
		"MALACHI":         4,
		"MATTHEW":         28,
		"MARK":            16,
		"LUKE":            24,
		"JOHN":            21,
		"ACTS":            28,
		"ROMANS":          16,
		"1CORINTHIANS":    16,
		"2CORINTHIANS":    13,
		"GALATIANS":       6,
		"EPHESIANS":       6,
		"PHILIPPIANS":     4,
		"COLOSSIANS":      4,
		"1THESSALONIANS":  5,
		"2THESSALONIANS":  3,
		"1TIMOTHY":        6,
		"2TIMOTHY":        4,
		"TITUS":           3,
		"PHILEMON":        1,
		"HEBREWS":         13,
		"JAMES":           5,
		"1PETER":          5,
		"2PETER":          3,
		"1JOHN":           5,
		"2JOHN":           1,
		"3JOHN":           1,
		"JUDE":            1,
		"REVELATION":      22,
	}
)

type App struct {
	Router   *mux.Router
	Database *sql.DB
}

type Verse struct {
	Book    string `json:"Book"`
	Chapter int    `json:"Chapter"`
	Verse   int    `json:"Verse"`
	Text    string `json:"Text"`
}

func (app *App) SetupRouter() {
	app.Router.HandleFunc("/bible/search", app.search)
	app.Router.HandleFunc("/bible/random_verse", app.getRandomVerse)
	app.Router.HandleFunc("/bible/list_books/", app.listBooks)
	app.Router.HandleFunc("/bible/list_books", app.listBooks) // why do i have to be explicit about the post slash here..

	t := app.Router.PathPrefix("/bible/list_chapters").Subrouter()
	t.HandleFunc("/{book}", app.listChapters)

	s := app.Router.PathPrefix("/bible").Subrouter()
	s.HandleFunc("/{book}", app.getBook)
	s.HandleFunc("/{book}/{chapter}", app.getChapter)
	s.HandleFunc("/{book}/{chapter}/{verse}", app.getVerse)

}

const booksButtonsTemplate = `
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
      <p><button class="block" onclick="window.location.href= '{{ createLink $value }}';" >{{ $value }}</button></p>
      {{ end }}
   </body>
</html>
`

func (app *App) listBooks(w http.ResponseWriter, r *http.Request) {

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
		return fmt.Sprintf("%s?json=false", b)
	}}

	t, err := template.New("listBooks").Funcs(funcs).Parse(booksButtonsTemplate)
	if err != nil {
		fmt.Printf("Could not list books: %s\n", err)
	}

	booksStruct := struct {
		Books []string
		Color string
	}{
		Books: books,
		Color: kjv.GetRandomColor(),
	}

	// Return json response if requested
	if wantsJson(r) {
		jsonizeResponse(booksStruct, w, r)
		return
	}

	t.Execute(w, booksStruct)
}

const chapterButtonsTemplate = `
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
       <p><button class="block" onclick="window.location.href = '{{ $results }}'">{{ add $index 1 }}</button></p>
     {{ end }}
   </body>
</html>
`

func (app *App) getBook(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var chapters struct {
		Links []string
		Name  string
		Color string
	}
	chapters.Name = strings.ToUpper(vars["book"])
	chapters.Color = kjv.GetRandomColor()

	// Handle non existant book
	chapterSize := BookChapterLimit[chapters.Name]
	if chapterSize == 0 {
		w.WriteHeader(http.StatusNotAcceptable)
		msg := fmt.Sprintf("406 - %s does not exist", vars["book"])
		w.Write([]byte(msg))
		return
	}

	// Create the chapter list
	for i := 1; i <= BookChapterLimit[chapters.Name]; i++ {
		link := fmt.Sprintf("%s/%d", chapters.Name, i)
		chapters.Links = append(chapters.Links, link)
	}

	// If json requested..
	if wantsJson(r) {
		jsonizeResponse(chapters, w, r)
		return
	}

	// Define the template func for add
	funcs := template.FuncMap{"add": func(x, y int) int { return x + y }}
	t, err := template.New("chapters").Funcs(funcs).Parse(chapterButtonsTemplate)
	if err != nil {
		panic(err)
	}

	t.Execute(w, chapters)
}

// ListChapters list the chapters of the book with clickable buttons for navigation
func (app *App) listChapters(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)

	// check arg is not empty
	var bookFound bool

	if vars["book"] != "" {
		vars["book"] = strings.ToUpper(vars["book"])
		//check the book actually exists
		for book, _ := range BookChapterLimit {
			if vars["book"] == book {
				bookFound = true
				break
			}
		}

		// Book not found..
		if !bookFound {
			w.WriteHeader(http.StatusNotAcceptable)
			msg := fmt.Sprintf("406 - %s does not exist", vars["book"])
			w.Write([]byte(msg))
			return
		}

	}

	chaptersMax := BookChapterLimit[strings.ToUpper(vars["book"])]

	chapterInfo := struct {
		Name     string
		Chapters []int
		Links    []string
		Color    string
	}{
		Name:  strings.ToUpper(vars["book"]),
		Color: kjv.GetRandomColor(),
	}

	// populate the data by interation
	for i := 1; i <= chaptersMax; i++ {
		chapterInfo.Chapters = append(chapterInfo.Chapters, i)
		chapterInfo.Links = append(
			chapterInfo.Links,
			fmt.Sprintf("bible/%s/%s?json=false", chapterInfo.Name, strconv.Itoa(i)))
	}

	// Return json response if requested
	if wantsJson(r) {
		jsonizeResponse(chapterInfo, w, r)
		return
	}

	funcs := template.FuncMap{"add": func(x, y int) int { return x + y }}
	t, err := template.New("chapterListing").Funcs(funcs).Parse(chapterButtonsTemplate)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("500 - Could not parse chapterButtonsTemplate"))
		return

	}
	fmt.Printf("%v\n", chapterInfo)
	w.Header().Set("Content-Type", "text/html")
	t.Execute(w, chapterInfo)
}

// getRandomVerseFromDB gets the verse from db to pass to pretty print api.
func (app *App) getRandomVerseFromDB() (Verse, error) {

	var randVerse Verse

	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)

	stmt := fmt.Sprintf("select book, chapter, verse, text from kjv where ordinal_verse=%d",
		r1.Intn(lastCardinalVerseNum))

	rows, err := app.Database.Query(stmt)
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

func (app *App) search(w http.ResponseWriter, r *http.Request) {

	var matches struct {
		Verses []Verse
		Count  map[string]int
	}

	var defaultSearchLimit = "10000"

	// Handle text query
	searchText, ok := r.URL.Query()["q"]
	fmt.Printf("%v\n", searchText)
	if !ok || len(searchText) < 1 {
		w.Write([]byte("Ye ask, and receive not, because ye ask amiss, that ye may consume it upon your lusts."))
		return
	}

	// Handle limit size
	searchLimit, ok := r.URL.Query()["n"]
	fmt.Printf("searchLImit initial: %v\n", searchLimit)
	if !ok || len(searchLimit) < 1 {
		searchLimit = append(searchLimit, defaultSearchLimit)
	}

	limit, err := strconv.Atoi(searchLimit[0])
	if err != nil {
		fmt.Println("Whoopsi with the limit size.")
		w.WriteHeader(http.StatusNotAcceptable)
		w.Write([]byte("whoopsie with the limit size.."))
		return
	}

	rows, err := app.Database.Query("select book, chapter, verse, text from kjv where text like ? limit ?", "%"+searchText[0]+"%", limit)
	if err != nil {
		w.Header().Set("Content-Type", "application/text")
		w.WriteHeader(http.StatusInternalServerError)
		msg := fmt.Sprintf("Failed to query: %s\n")
		w.Write([]byte(msg))
		return
		log.Println(err)
	}

	regexCount := 0
	overallCount := make(map[string]int)
	re := regexp.MustCompile("(?i)" + searchText[0])

	for rows.Next() {
		match := Verse{}
		err := rows.Scan(&match.Book, &match.Chapter, &match.Verse, &match.Text)

		if err != nil {
			w.Header().Set("Content-Type", "application/text")
			w.WriteHeader(http.StatusInternalServerError)
			msg := fmt.Sprintf("Failed to scan query: %s\n", err)
			w.Write([]byte(msg))
			return
		}

		//////////////////////////////
		// Count regex finds	    //
		//////////////////////////////
		foundCount := re.FindAll([]byte(match.Text), -1)
		regexCount = regexCount + len(foundCount)

		overallCount[match.Book] += 1
		overallCount["overall"] += 1
		matches.Verses = append(matches.Verses, match)
	}

	matches.Count = overallCount
	// Handle json request
	if wantsJson(r) {
		jsonizeResponse(matches, w, r)
		return
	}

	// template func to create href
	funcs := template.FuncMap{"createLink": func(a Verse) string {
		return strings.Join([]string{
			a.Book,
			strconv.Itoa(a.Chapter),
			strconv.Itoa(a.Verse),
		},
			"/")
	}}

	const searchResultTemplate = `
<!DOCTYPE html>
<html>

      {{range .Verses }}
	 <center>   
           <p> <a href="{{ createLink .}}">{{ .Book }} {{ .Chapter }}:{{ .Verse}} </a></p>
           <p> {{ .Text }} </p>
         </center>
      {{ end }}

   </body>
</html>
`
	tmpl, err := template.New("results").Funcs(funcs).Parse(searchResultTemplate)
	if err != nil {
		fmt.Println("Failed to parse template..")
		return
	}
	err = tmpl.Execute(w, matches)
	if err != nil {
		w.Header().Set("Content-Type", "application/text")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to write to Writer"))
		log.Println(err)
		return
	}

}

func jsonizeResponse(obj interface{}, w http.ResponseWriter, r *http.Request) {

	jsonResult, err := json.MarshalIndent(obj, "  ", "")
	if err != nil {
		w.Header().Set("Content-Type", "application/text")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to marshal json from result"))
		log.Println(err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResult)
	return

}

// getRandomVerse write pretty html page with random verse.
func (app *App) getRandomVerse(w http.ResponseWriter, r *http.Request) {

	result, err := app.getRandomVerseFromDB()
	if err != nil {

		w.Header().Set("Content-Type", "application/text")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to get random verse from db."))
		return
	}

	// create the data struct for template
	returnPage := struct {
		Verse      Verse
		Color      string
		ChapterRef string
	}{
		result,
		kjv.GetRandomColor(),
		"",
	}

	// Return json response if requested
	if wantsJson(r) {
		jsonizeResponse(result, w, r)
		return
	}

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
	returnPage.ChapterRef = fmt.Sprintf("%s/%d?json=false",
		result.Book,
		result.Chapter,
	)

	err = tmpl.Execute(w, returnPage)
	if err != nil {
		w.Header().Set("Content-Type", "application/text")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to write to Writer"))
		log.Println(err)
		return
	}
}

func wantsJson(r *http.Request) bool {
	if r.URL.Query()["json"] != nil {
		if r.URL.Query()["json"][0] == "true" {
			return true
		}
	}

	return false
}

func (app *App) getChapter(w http.ResponseWriter, r *http.Request) {

	var (
		verses = struct {
			BookName            string
			Chapter             int
			Verses              []string
			Color               string
			NextChapterLink     string
			PreviousChapterLink string
			ListAllBooksLink    string
		}{
			Color:            kjv.GetRandomColor(),
			ListAllBooksLink: "../list_books?json=false",
		}

		vars = mux.Vars(r)

		// check arg is not empty
		bookFound bool
	)

	if vars["book"] != "" {
		vars["book"] = strings.ToUpper(vars["book"])
		//check the book actually exists
		for book, _ := range BookChapterLimit {
			if vars["book"] == book {
				bookFound = true
				break
			}
		}

		// Book not found..
		if !bookFound {
			w.WriteHeader(http.StatusNotAcceptable)
			msg := fmt.Sprintf("406 - %s does not exist", vars["book"])
			w.Write([]byte(msg))
			return
		}

	}

	verses.BookName = vars["book"]

	chapter, err := strconv.Atoi(vars["chapter"])
	if err != nil {
		w.WriteHeader(http.StatusNotAcceptable)
		w.Write([]byte(fmt.Sprintf("%s is not a proper chapter", vars["chapter"])))
		return
	}

	// check the chapter is not > last chapter for book
	if chapter > BookChapterLimit[verses.BookName] {
		w.WriteHeader(http.StatusNotAcceptable)
		msg := fmt.Sprintf("Chapter %d is out of bounds, last chapter of %s is %d\n", chapter, verses.BookName, BookChapterLimit[verses.BookName])
		w.Write([]byte(msg))
		return
	}

	verses.Chapter = chapter

	stmt := fmt.Sprintf("select verse, text from kjv where book='%s' and chapter=%v", verses.BookName, verses.Chapter)

	rows, err := app.Database.Query(stmt)
	defer rows.Close()

	if err != nil {
		log.Println(err)
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

	// Add footer next chapter and previous chapter
	if verses.Chapter <= 1 {
		verses.PreviousChapterLink = ""
	} else {
		verses.PreviousChapterLink = fmt.Sprintf("%s?json=false", strconv.Itoa(verses.Chapter-1))
	}

	if verses.Chapter < BookChapterLimit[verses.BookName] {
		// verses.NextChapterLink = fmt.Sprintf("get_chapter?book=%s&chapter=%s", verses.BookName, strconv.Itoa(verses.Chapter+1))
		verses.NextChapterLink = fmt.Sprintf("%s?json=false", strconv.Itoa(verses.Chapter+1))
	}

	// Return json response if requested
	if wantsJson(r) {
		jsonizeResponse(verses, w, r)
		return
	}

	//////////////////////////
	// Template time        //
	//////////////////////////
	// add function to increment range indexing since it starts at 0 by default
	funcs := template.FuncMap{"add": func(x, y int) int { return x + y }}
	verseLink := template.FuncMap{"verseLink": func(x int) string {

		verseOffSet := strconv.Itoa(x + 1)

		return fmt.Sprintf("%s/%s?json=false",
			strconv.Itoa(verses.Chapter),
			verseOffSet,
		)
	}}

	t, err := template.New("chapter").Funcs(funcs).Funcs(verseLink).Parse(chapterTemplate)

	if err != nil {
		panic(err)
	}

	t.Execute(w, verses)
}

func (app *App) getVerse(w http.ResponseWriter, r *http.Request) {
	var (
		verse       Verse
		bookFound   bool
		requestVars = mux.Vars(r)
	)

	//check the book actually exists
	if requestVars["book"] != "" {
		requestVars["book"] = strings.ToUpper(requestVars["book"])
		for book, _ := range BookChapterLimit {
			if requestVars["book"] == book {
				bookFound = true
				break
			}
		}

		// Book not found..
		if !bookFound {
			w.WriteHeader(http.StatusNotAcceptable)
			msg := fmt.Sprintf("406 - %s does not exist", requestVars["book"])
			w.Write([]byte(msg))
			return
		}
	}

	verse.Book = requestVars["book"]

	// Check Chapter
	rChapter, err := strconv.Atoi(requestVars["chapter"])
	if err != nil {
		msg := fmt.Sprintf("%s is not an integer chapter\n", requestVars["chapter"])
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	verse.Chapter = rChapter

	// Check Verse
	rVerse, err := strconv.Atoi(requestVars["verse"])
	if err != nil {
		msg := fmt.Sprintf("%s is not an integer verse reference\n", requestVars["chapter"])
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	verse.Verse = rVerse

	// Query the database
	stmt := fmt.Sprintf("select text from kjv where book=\"%s\" and chapter=%s and verse=%s",
		requestVars["book"],
		strconv.Itoa(rChapter),
		strconv.Itoa(rVerse),
	)

	rows, err := app.Database.Query(stmt)
	if err != nil {
		http.Error(w, fmt.Sprintf("Could not query DB"), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		rows.Scan(&verse.Text)
	}

	// Check results from DB
	if len(verse.Text) <= 0 {
		msg := fmt.Sprintf("Got nothing back from database: %s", stmt)
		http.Error(w, msg, http.StatusInternalServerError)

	}

	// Return json response if requested
	if wantsJson(r) {
		jsonizeResponse(verse, w, r)
		return
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
		Verse      Verse
		Color      string
		ChapterRef string
	}{
		Verse:      verse,
		Color:      kjv.GetRandomColor(),
		ChapterRef: fmt.Sprintf("../%s", strconv.Itoa(verse.Chapter)),
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

}

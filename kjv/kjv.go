package kjv

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"text/template"

	"net/http"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	kjv "github.com/r4wm/bible_api"
)

const lastCardinalVerseNum = 31101

// BooksCanonicalOrder is the 66 books of the Bible in their traditional order.
var BooksCanonicalOrder = []string{
	"GENESIS", "EXODUS", "LEVITICUS", "NUMBERS", "DEUTERONOMY",
	"JOSHUA", "JUDGES", "RUTH", "1SAMUEL", "2SAMUEL",
	"1KINGS", "2KINGS", "1CHRONICLES", "2CHRONICLES",
	"EZRA", "NEHEMIAH", "ESTHER", "JOB", "PSALMS", "PROVERBS",
	"ECCLESIASTES", "SONG OF SOLOMON", "ISAIAH", "JEREMIAH", "LAMENTATIONS",
	"EZEKIEL", "DANIEL", "HOSEA", "JOEL", "AMOS",
	"OBADIAH", "JONAH", "MICAH", "NAHUM", "HABAKKUK",
	"ZEPHANIAH", "HAGGAI", "ZECHARIAH", "MALACHI",
	"MATTHEW", "MARK", "LUKE", "JOHN", "ACTS",
	"ROMANS", "1CORINTHIANS", "2CORINTHIANS", "GALATIANS", "EPHESIANS",
	"PHILIPPIANS", "COLOSSIANS", "1THESSALONIANS", "2THESSALONIANS",
	"1TIMOTHY", "2TIMOTHY", "TITUS", "PHILEMON", "HEBREWS",
	"JAMES", "1PETER", "2PETER", "1JOHN", "2JOHN", "3JOHN",
	"JUDE", "REVELATION",
}

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
	Router *mux.Router
	Redis  *redis.Client

	OpenSearchURL      string
	OpenSearchIndex    string
	OpenSearchUsername string
	OpenSearchPassword string
	OpenSearchHTTP     *http.Client

	// Immutable after startup. Preloaded from OpenSearch.
	AllVerses  map[string]map[int][]Verse // book → chapter → sorted verses
	FlatVerses []Verse                    // flat slice for random access

	JWTSecret           []byte
	JWTIssuer           string
	JWTAudience         string
	JWTTTLSeconds       int
	SessionTTLSeconds   int
	SessionCookieName   string
	SessionCookieSecure bool
	GoogleClientID      string
	InternalTokenSecret string
	StripeSecretKey     string
	StripeAPIBaseURL    string
	StripeHTTP          *http.Client
	PublicBaseURL       string
}

type Verse struct {
	Book    string `json:"Book"`
	Chapter int    `json:"Chapter"`
	Verse   int    `json:"Verse"`
	Text    string `json:"Text"`
}

func (v *Verse) RemoveItalicMarkers() {
	v.Text = strings.ReplaceAll(v.Text, "[", "")
	v.Text = strings.ReplaceAll(v.Text, "]", "")
}

// Unified JSON Response Structures
type UnifiedResponse struct {
	Status string       `json:"status"`
	Meta   ResponseMeta `json:"meta"`
	Data   ResponseData `json:"data"`
}

type ResponseMeta struct {
	Query       string `json:"query,omitempty"`
	Count       int    `json:"count"`
	Timestamp   string `json:"timestamp"`
	QueryType   string `json:"query_type,omitempty"`
	SearchLimit int    `json:"search_limit,omitempty"`
}

type ResponseData struct {
	Books []BookData `json:"books"`
}

type BookData struct {
	Name         string        `json:"name"`
	ChapterCount int           `json:"chapter_count,omitempty"`
	Chapters     []ChapterData `json:"chapters,omitempty"`
}

type ChapterData struct {
	Number     int         `json:"number"`
	VerseCount int         `json:"verse_count,omitempty"`
	Verses     []VerseData `json:"verses,omitempty"`
}

type VerseData struct {
	Number int    `json:"number"`
	Text   string `json:"text"`
}

func (app *App) SetupRouter() {
	app.Router.HandleFunc("/bible/search", app.search)
	app.Router.HandleFunc("/bible/suggest", app.suggestCompletion).Methods("GET")
	app.Router.HandleFunc("/bible/random_verse", app.getRandomVerse)
	app.Router.HandleFunc("/bible/list_books/", app.listBooks)
	app.Router.HandleFunc("/bible/list_books", app.listBooks) // why do i have to be explicit about the post slash here..

	t := app.Router.PathPrefix("/bible/list_chapters").Subrouter()
	t.HandleFunc("/{book}", app.listChapters)

	// Register v2 routes before v1 /bible/{book} routes so they don't get shadowed.
	app.SetupV2Routes()

	s := app.Router.PathPrefix("/bible").Subrouter()
	s.HandleFunc("/{book}", app.getBook)
	s.HandleFunc("/{book}/{chapter}", app.getChapter)
	s.HandleFunc("/{book}/{chapter}/{verse}", app.getVerse)

	// Setup admin routes for rate limit management
	app.SetupAdminRoutes()
	app.SetupAuthRoutes()
	app.SetupUserRoutes()
	app.SetupDocsRoutes()
	app.SetupDonationRoutes()
	app.SetupUIRoutes()
}

func (app *App) listBooks(w http.ResponseWriter, r *http.Request) {
	books := BooksCanonicalOrder
	// funcs generates the link needed for button
	funcs := template.FuncMap{"createLink": func(b string) string {
		return fmt.Sprintf("%s?json=false", b)
	}}

	t, err := template.New("listBooks").Funcs(funcs).Parse(booksButtonsTemplate)
	if err != nil {
		http.Error(w, "Could not parse template", http.StatusInternalServerError)
		log.Printf("Template parsing error: %v", err)
		return
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
		jsonizeResponse(booksStruct, w)
		return
	}

	if err := t.Execute(w, booksStruct); err != nil {
		http.Error(w, "Could not execute template", http.StatusInternalServerError)
		log.Printf("Template execution error: %v", err)
	}
}

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
		jsonizeResponse(chapters, w)
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
		for book := range BookChapterLimit {
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
		jsonizeResponse(chapterInfo, w)
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

// getRandomVerseFromDB returns a random verse from the in-memory cache, or falls back to OpenSearch.
func (app *App) getRandomVerseFromDB() (Verse, error) {
	if len(app.FlatVerses) > 0 {
		return app.FlatVerses[rand.Intn(len(app.FlatVerses))], nil
	}
	return app.osGetRandomVerse()
}

func (app *App) search(w http.ResponseWriter, r *http.Request) {

	var matches struct {
		Verses       []Verse
		SearchString string
		Count        map[string]int
		GraphCount   string // json array of ints
	}

	// Handle text query
	searchText, ok := r.URL.Query()["q"]
	if !ok || len(searchText) < 1 {
		w.Write([]byte("Ye ask, and receive not, because ye ask amiss, that ye may consume it upon your lusts."))
		return
	}

	// Handle limit size
	limit := 10000
	if n := r.URL.Query().Get("n"); n != "" {
		if parsed, err := strconv.Atoi(n); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	// Check if show_italics parameter is present
	showItalics := r.URL.Query().Get("show_italics") == "true"

	matches.SearchString = searchText[0]

	// Query OpenSearch
	verses, graphCounter, err := app.osSearch(searchText[0], limit)
	if err != nil {
		http.Error(w, "Search service unavailable", http.StatusServiceUnavailable)
		log.Printf("OpenSearch search error: %v", err)
		return
	}

	// Strip italic markers unless show_italics is set
	if !showItalics {
		for i := range verses {
			verses[i].RemoveItalicMarkers()
		}
	}

	matches.Verses = verses

	// Build per-book count map for JSON response
	overallCount := make(map[string]int)
	for _, v := range verses {
		overallCount[v.Book]++
		overallCount["overall"]++
	}
	matches.Count = overallCount

	// Handle json request
	if wantsJson(r) {
		jsonizeResponse(matches, w)
		return
	}

	// template func to create href
	funcs := template.FuncMap{"createLink": func(a Verse) string {
		return strings.Join([]string{
			a.Book,
			strconv.Itoa(a.Chapter),
			strconv.Itoa(a.Verse),
		}, "/")
	}}

	tmpl, err := template.New("results").Funcs(funcs).Parse(searchResultTemplate)
	if err != nil {
		http.Error(w, "Failed to parse search template", http.StatusInternalServerError)
		log.Printf("Template parsing error: %v", err)
		return
	}

	graphBytes, err := json.Marshal(graphCounter)
	if err != nil {
		http.Error(w, "Failed to marshal graph data", http.StatusInternalServerError)
		log.Printf("JSON marshaling error: %v", err)
		return
	}

	matches.GraphCount = string(graphBytes)

	if err := tmpl.Execute(w, matches); err != nil {
		http.Error(w, "Failed to execute search template", http.StatusInternalServerError)
		log.Printf("Template execution error: %v", err)
	}
}

func jsonizeResponse(obj interface{}, w http.ResponseWriter) {
	jsonResult, err := json.MarshalIndent(obj, "  ", "")
	if err != nil {
		http.Error(w, "Failed to marshal JSON response", http.StatusInternalServerError)
		log.Printf("JSON marshaling error: %v", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResult)
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
		jsonizeResponse(result, w)
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

// lazyBook only adds book to string if pattern is met
func lazyBook(shortName string) (book string, err error) {
	// TODO: write simple test
	shortName = strings.ToUpper(shortName)
	var possibleBooks []string

	// iterate the BookList and add if pattern meets
	for bookCandidate := range BookChapterLimit {
		// fmt.Println("this is candidate: ", bookCandidate)
		// TODO CHange to regex /
		if strings.HasPrefix(bookCandidate, shortName) {
			// fmt.Println("Found : ", bookCandidate)
			possibleBooks = append(possibleBooks, bookCandidate)
		}
	}

	if len(possibleBooks) > 1 {
		errMsg := fmt.Sprintf("more than one possible choice: %s", possibleBooks)
		err = errors.New(errMsg)
		return "", err
	} else if len(possibleBooks) == 1 {
		book = possibleBooks[0]
		return book, nil
	}

	return "", errors.New("no matching book found")
}

func (app *App) getChapter(w http.ResponseWriter, r *http.Request) {
	fmt.Println("calling getChapter")
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
	)

	// Check if show_italics parameter is present and set to true
	showItalics := false
	if italicsParam := r.URL.Query().Get("show_italics"); italicsParam == "true" {
		showItalics = true
	}
	fmt.Printf("show italics: is %v\n", showItalics)

	book := strings.ToUpper(vars["book"])
	bookName, err := lazyBook(book)
	if err != nil {
		w.WriteHeader(http.StatusNotAcceptable)
		w.Write([]byte(fmt.Sprintf("%s", err)))
		return
	}
	verses.BookName = bookName

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

	var osVerses []Verse
	if app.AllVerses != nil {
		if chapters, ok := app.AllVerses[verses.BookName]; ok {
			osVerses = chapters[verses.Chapter]
		}
	} else {
		var osErr error
		osVerses, osErr = app.osGetChapter(verses.BookName, verses.Chapter)
		if osErr != nil {
			http.Error(w, "Search service unavailable", http.StatusServiceUnavailable)
			log.Printf("OpenSearch getChapter error: %v", osErr)
			return
		}
	}

	for _, v := range osVerses {
		text := v.Text
		if !showItalics {
			text = strings.ReplaceAll(text, "[", "")
			text = strings.ReplaceAll(text, "]", "")
		}
		verses.Verses = append(verses.Verses, text)
	}

	// Add footer next chapter and previous chapter
	if verses.Chapter <= 1 {
		verses.PreviousChapterLink = ""
	} else {
		verses.PreviousChapterLink = fmt.Sprintf("%s?json=false", strconv.Itoa(verses.Chapter-1))
	}

	if verses.Chapter < BookChapterLimit[verses.BookName] {
		verses.NextChapterLink = fmt.Sprintf("%s?json=false", strconv.Itoa(verses.Chapter+1))
	}

	// Return json response if requested
	if wantsJson(r) {
		jsonizeResponse(verses, w)
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
		verses = struct {
			HTMLTitle           string
			BookName            string
			Chapter             int
			Verses              []map[int]string
			Color               string
			NextChapterLink     string
			PreviousChapterLink string
			ListAllBooksLink    string
			StartVerse          int
			EndVerse            int
			SingleVerse         int
		}{
			Color:            kjv.GetRandomColor(),
			ListAllBooksLink: "../../list_books?json=false",
		}
		requestVars = mux.Vars(r)
		bookFound   bool
	)

	//check the book actually exists
	bookName, err := lazyBook(requestVars["book"])
	if err != nil {
		w.WriteHeader(http.StatusNotAcceptable)
		w.Write([]byte(fmt.Sprintf("%s", err)))
		return
	}

	if bookName != "" {
		bookName = strings.ToUpper(bookName)
		for book := range BookChapterLimit {
			if bookName == book {
				bookFound = true
				break
			}
		}

		// Book not found..
		if !bookFound {
			w.WriteHeader(http.StatusNotAcceptable)
			msg := fmt.Sprintf("406 - %s does not exist", bookName)
			w.Write([]byte(msg))
			return
		}
	}

	verses.BookName = bookName

	// Check Chapter
	rChapter, err := strconv.Atoi(requestVars["chapter"])
	if err != nil {
		msg := fmt.Sprintf("Chapter %s is not an integer chapter\n", requestVars["chapter"])
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	verses.Chapter = rChapter

	// Check Verse
	isVerseRange := strings.Contains(requestVars["verse"], "-")

	var verseNums []int

	if isVerseRange {
		// Multiple Verse
		log.Printf("Checking for valid range: %s", requestVars["verse"])
		verseRange := strings.Split(requestVars["verse"], "-")

		verses.StartVerse, err = strconv.Atoi(verseRange[0])
		if err != nil {
			http.Error(w, fmt.Sprintf("Verse range start is not valid: %s", verseRange[0]), http.StatusBadRequest)
			return
		}

		verses.EndVerse, err = strconv.Atoi(verseRange[1])
		if err != nil {
			http.Error(w, "Verse range end is not valid: "+verseRange[1], http.StatusBadRequest)
			return
		}

		for i := verses.StartVerse; i <= verses.EndVerse; i++ {
			verseNums = append(verseNums, i)
		}

		verses.HTMLTitle = fmt.Sprintf("%s %s:%s-%s",
			bookName,
			strconv.Itoa(rChapter),
			verseRange[0],
			verseRange[1],
		)

	} else {
		// Single verse
		rVerse, err := strconv.Atoi(requestVars["verse"])
		if err != nil {
			msg := fmt.Sprintf("Verse %s is not an integer verse reference\n", requestVars["chapter"])
			http.Error(w, msg, http.StatusBadRequest)
			return
		}

		verses.SingleVerse = rVerse
		verseNums = []int{rVerse}

		verses.HTMLTitle = fmt.Sprintf("%s %s:%s",
			bookName,
			strconv.Itoa(rChapter),
			requestVars["verse"],
		)
	}

	var osVerses []Verse
	if app.AllVerses != nil {
		if chapters, ok := app.AllVerses[bookName]; ok {
			wanted := make(map[int]bool, len(verseNums))
			for _, n := range verseNums {
				wanted[n] = true
			}
			for _, v := range chapters[rChapter] {
				if wanted[v.Verse] {
					osVerses = append(osVerses, v)
				}
			}
		}
	} else {
		var osErr error
		osVerses, osErr = app.osGetVerses(bookName, rChapter, verseNums)
		if osErr != nil {
			http.Error(w, "Search service unavailable", http.StatusServiceUnavailable)
			log.Printf("OpenSearch getVerses error: %v", osErr)
			return
		}
	}

	for _, v := range osVerses {
		verses.Verses = append(verses.Verses, map[int]string{v.Verse: v.Text})
	}

	if wantsJson(r) {
		jsonizeResponse(verses, w)
		return
	}

	////////////////////////////
	// Create Template	  //
	////////////////////////////
	t, err := template.New("chapter").Parse(versesTemplate)

	if err != nil {
		panic(err)
	}

	t.Execute(w, verses)
}

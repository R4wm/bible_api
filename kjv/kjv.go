package kjv

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	kjv "github.com/r4wm/bible_api"
	log "github.com/sirupsen/logrus"
)

const lastCardinalVerseNum = 31101

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
	Redis    *redis.Client
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
func (app *App) SetupRouter() {
	app.Router.HandleFunc("/bible/search", app.search)
	app.Router.HandleFunc("/bible/autocomplete", app.autocomplete)
	app.Router.HandleFunc("/bible/random_verse", app.getRandomVerse)
	app.Router.HandleFunc("/bible/list_books/", app.listBooks)
	app.Router.HandleFunc("/bible/list_books", app.listBooks) // why do i have to be explicit about the post slash here..
	app.Router.HandleFunc("/bible/daily/proverbs", app.GetDailyProverbs)
	app.Router.HandleFunc("/bible/daily/psalms", app.GetDailyPsalms)
	app.Router.HandleFunc("/bible/daily/ot", app.GetDailyOldTestament)
	app.Router.HandleFunc("/bible/daily/nt", app.GetDailyNewTestament)
	// app.Router.HandleFunc("/bible/daily", app.getDaily)

	t := app.Router.PathPrefix("/bible/list_chapters").Subrouter()
	t.HandleFunc("/{book}", app.listChapters)

	s := app.Router.PathPrefix("/bible").Subrouter()
	s.HandleFunc("/{book}", app.getBook)
	s.HandleFunc("/{book}/{chapter}", app.getChapter)
	s.HandleFunc("/{book}/{chapter}/{verse}", app.getVerse)
	// TODO: Make this clean , reusable based on book
	s.HandleFunc("/daily/proverbs", app.GetDailyProverbs)
	s.HandleFunc("/daily/psalms", app.GetDailyPsalms)
	s.HandleFunc("/daily/ot", app.GetDailyOldTestament)
	s.HandleFunc("/daily/nt", app.GetDailyNewTestament)
	// s.HandleFunc("/daily", app.getDaily)

	// Setup admin routes for rate limit management
	app.SetupAdminRoutes()
}

func (app *App) listBooks(w http.ResponseWriter, r *http.Request) {
	// Maintain proper biblical order instead of random map iteration
	books := []string{
		"GENESIS", "EXODUS", "LEVITICUS", "NUMBERS", "DEUTERONOMY",
		"JOSHUA", "JUDGES", "RUTH", "1SAMUEL", "2SAMUEL",
		"1KINGS", "2KINGS", "1CHRONICLES", "2CHRONICLES", "EZRA",
		"NEHEMIAH", "ESTHER", "JOB", "PSALMS", "PROVERBS",
		"ECCLESIASTES", "SONG OF SOLOMON", "ISAIAH", "JEREMIAH", "LAMENTATIONS",
		"EZEKIEL", "DANIEL", "HOSEA", "JOEL", "AMOS",
		"OBADIAH", "JONAH", "MICAH", "NAHUM", "HABAKKUK",
		"ZEPHANIAH", "HAGGAI", "ZECHARIAH", "MALACHI", "MATTHEW",
		"MARK", "LUKE", "JOHN", "ACTS", "ROMANS",
		"1CORINTHIANS", "2CORINTHIANS", "GALATIANS", "EPHESIANS", "PHILIPPIANS",
		"COLOSSIANS", "1THESSALONIANS", "2THESSALONIANS", "1TIMOTHY", "2TIMOTHY",
		"TITUS", "PHILEMON", "HEBREWS", "JAMES", "1PETER",
		"2PETER", "1JOHN", "2JOHN", "3JOHN", "JUDE", "REVELATION",
	}

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
		http.Error(w, "Could not parse template", http.StatusInternalServerError)
		log.Printf("Template parsing error: %v", err)
		return
	}

	if err := t.Execute(w, chapters); err != nil {
		http.Error(w, "Could not execute template", http.StatusInternalServerError)
		log.Printf("Template execution error: %v", err)
	}
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
	log.Debugf("Chapter info: %+v", chapterInfo)
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
		log.Errorf("Failed DB.Query(%s): %v", stmt, err)
		return randVerse, err
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&randVerse.Book, &randVerse.Chapter, &randVerse.Verse, &randVerse.Text); err != nil {
			log.Errorf("Failed to scan random verse row: %v", err)
			return randVerse, err
		}
	}

	// OK
	return randVerse, nil
}

func (app *App) search(w http.ResponseWriter, r *http.Request) {

	var matches struct {
		Verses       []Verse
		SearchString string
		Count        map[string]int
		GraphCount   string // json array of ints
	}

	graphBookCounter := [66]int{}
	var defaultSearchLimit = "10000"

	// Handle text query
	searchText, ok := r.URL.Query()["q"]
	if !ok || len(searchText) < 1 {
		// Show search form instead of biblical quote
		app.showSearchForm(w, r)
		return
	}
	log.Debugf("Search query: %s", searchText[0])
	// Validate search string - allow only alphanumeric, spaces, and common punctuation
	if matched, _ := regexp.MatchString(`[^\w\s'".,;:!?()-]`, searchText[0]); matched {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Search string contains invalid characters"))
		return
	}

	// Prevent extremely short or long search terms
	if len(strings.TrimSpace(searchText[0])) < 2 || len(searchText[0]) > 100 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Search string must be between 2 and 100 characters"))
		return
	}

	// Handle limit size
	searchLimit, ok := r.URL.Query()["n"]
	var limitStr string
	if !ok || len(searchLimit) < 1 {
		limitStr = defaultSearchLimit
	} else {
		limitStr = searchLimit[0]
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 10000 {
		log.Warnf("Invalid search limit provided: %s", limitStr)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Search limit must be a number between 1 and 10000"))
		return
	}

	// Check if show_italics parameter is present and set to true
	showItalics := false
	if italicsParam := r.URL.Query().Get("show_italics"); italicsParam == "true" {
		showItalics = true
	}

	matches.SearchString = searchText[0]

	rows, err := app.Database.Query("select book, chapter, verse, text, ordinal_book from kjv where replace(replace(text, '[', ''), ']', '') like ? limit ?", "%"+searchText[0]+"%", limit)
	if err != nil {
		w.Header().Set("Content-Type", "application/text")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed database query!"))
		log.Println(err)
		return
	}

	regexCount := 0
	overallCount := make(map[string]int)
	re := regexp.MustCompile("(?i)" + searchText[0])

	for rows.Next() {
		match := Verse{}
		var ordinalBook int
		err := rows.Scan(&match.Book, &match.Chapter, &match.Verse, &match.Text, &ordinalBook)

		if !showItalics {
			match.RemoveItalicMarkers()
		}

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
		graphBookCounter[ordinalBook-1] += 1
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
		},
			"/")
	}}

	tmpl, err := template.New("results").Funcs(funcs).Parse(searchResultTemplate)
	if err != nil {
		http.Error(w, "Failed to parse search template", http.StatusInternalServerError)
		log.Printf("Template parsing error: %v", err)
		return
	}

	graphBytes, err := json.Marshal(graphBookCounter)
	if err != nil {
		http.Error(w, "Failed to marshal graph data", http.StatusInternalServerError)
		log.Printf("JSON marshaling error: %v", err)
		return
	}

	matches.GraphCount = string(graphBytes)

	err = tmpl.Execute(w, matches)
	if err != nil {
		http.Error(w, "Failed to execute search template", http.StatusInternalServerError)
		log.Printf("Template execution error: %v", err)
		return
	}

}

func (app *App) autocomplete(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	
	if len(query) < 2 {
		jsonizeResponse([]string{}, w)
		return
	}
	
	// Simple approach: get verses that contain words starting with the query
	rows, err := app.Database.Query(`
		SELECT DISTINCT text 
		FROM kjv 
		WHERE LOWER(text) LIKE '%' || LOWER(?) || '%'
		LIMIT 50
	`, query)
	
	if err != nil {
		log.Printf("Autocomplete query error: %v", err)
		jsonizeResponse([]string{}, w)
		return
	}
	defer rows.Close()
	
	// Extract words from the verse texts
	wordSet := make(map[string]bool)
	suggestions := []string{}
	
	for rows.Next() {
		var text string
		if err := rows.Scan(&text); err != nil {
			continue
		}
		
		// Split text into words and find matches
		words := strings.Fields(strings.ToLower(text))
		for _, word := range words {
			// Clean word of punctuation
			cleaned := strings.Trim(word, ".,;:!?()[]\"'")
			if len(cleaned) > 1 && strings.HasPrefix(cleaned, strings.ToLower(query)) {
				if !wordSet[cleaned] {
					wordSet[cleaned] = true
					suggestions = append(suggestions, cleaned)
					if len(suggestions) >= 10 {
						break
					}
				}
			}
		}
		if len(suggestions) >= 10 {
			break
		}
	}
	
	jsonizeResponse(suggestions, w)
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
	log.Debug("Processing getChapter request")
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
	log.Debugf("Show italics parameter: %v", showItalics)

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

	// Use parameterized query to prevent SQL injection
	stmt := "select verse, text from kjv where book=? and chapter=?"

	rows, err := app.Database.Query(stmt, verses.BookName, verses.Chapter)
	if err != nil {
		log.Errorf("Database query failed: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("500 - Could not query database"))
		return
	}
	defer rows.Close()

	var verse int
	var text string

	for rows.Next() {
		if err := rows.Scan(&verse, &text); err != nil {
			log.Errorf("Failed to scan verse row: %v", err)
			continue
		}
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
		// verses.NextChapterLink = fmt.Sprintf("get_chapter?book=%s&chapter=%s", verses.BookName, strconv.Itoa(verses.Chapter+1))
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
		http.Error(w, "Could not parse template", http.StatusInternalServerError)
		log.Printf("Template parsing error: %v", err)
		return
	}

	if err := t.Execute(w, verses); err != nil {
		http.Error(w, "Could not execute template", http.StatusInternalServerError)
		log.Printf("Template execution error: %v", err)
	}
}

func (app *App) showSearchForm(w http.ResponseWriter, r *http.Request) {
	// Check if JSON is requested
	if wantsJson(r) {
		response := map[string]string{
			"message": "Please provide a search query parameter 'q'",
			"example": "/bible/search?q=love",
		}
		w.Header().Set("Content-Type", "application/json")
		jsonizeResponse(response, w)
		return
	}

	// Show HTML search form
	searchFormData := struct {
		Color string
	}{
		Color: kjv.GetRandomColor(),
	}

	tmpl, err := template.New("searchForm").Parse(searchFormTemplate)
	if err != nil {
		http.Error(w, "Could not parse search form template", http.StatusInternalServerError)
		log.Printf("Template parsing error: %v", err)
		return
	}

	if err := tmpl.Execute(w, searchFormData); err != nil {
		http.Error(w, "Could not execute search form template", http.StatusInternalServerError)
		log.Printf("Template execution error: %v", err)
	}
}
func (app *App) GetDailyProverbs(w http.ResponseWriter, r *http.Request) {

	versesFromProverbs := []Verse{}

	proverbsReading := GetProverbsDailyRange(GetDaysInMonth(), time.Now().Day())
	log.Debugf("Proverbs reading range: %+v", proverbsReading)

	stmt := "select book, chapter, verse, text from kjv where ordinal_verse between ? and ?"
	log.Debugf("Proverbs query: ordinal_verse between %d and %d", proverbsReading.StartOrdinalVerse, proverbsReading.EndOrdinalVerse)

	rows, err := app.Database.Query(stmt, proverbsReading.StartOrdinalVerse, proverbsReading.EndOrdinalVerse)
	if err != nil {
		log.Errorf("Failed to query database for daily proverbs: %v", err)
		http.Error(w, "Failed to retrieve daily proverbs", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		v := Verse{}
		if err := rows.Scan(&v.Book, &v.Chapter, &v.Verse, &v.Text); err != nil {
			log.Errorf("Failed to scan proverbs verse: %v", err)
			continue
		}
		versesFromProverbs = append(versesFromProverbs, v)
	}

	// TODO: Render HTML response , just JSON for now cause time
	jsonizeResponse(versesFromProverbs, w)
}

func (app *App) GetDailyPsalms(w http.ResponseWriter, r *http.Request) {

	versesFromPsalms := []Verse{}

	proverbsReading := GetPsalmsDailyRange(GetDaysInMonth(), time.Now().Day())
	log.Debugf("Psalms reading range: %+v", proverbsReading)

	stmt := "select book, chapter, verse, text from kjv where ordinal_verse between ? and ?"
	log.Debugf("Psalms query: ordinal_verse between %d and %d", proverbsReading.StartOrdinalVerse, proverbsReading.EndOrdinalVerse)

	rows, err := app.Database.Query(stmt, proverbsReading.StartOrdinalVerse, proverbsReading.EndOrdinalVerse)
	if err != nil {
		log.Errorf("Failed to query database for daily psalms: %v", err)
		http.Error(w, "Failed to retrieve daily psalms", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	for rows.Next() {
		v := Verse{}
		if err := rows.Scan(&v.Book, &v.Chapter, &v.Verse, &v.Text); err != nil {
			log.Errorf("Failed to scan psalms verse: %v", err)
			continue
		}
		versesFromPsalms = append(versesFromPsalms, v)
	}

	// TODO: Render HTML response , just JSON for now cause time
	jsonizeResponse(versesFromPsalms, w)
}

func (app *App) GetDailyOldTestament(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Old Testament Daily Range")
	versesFromOT := []Verse{}

	t := time.Now()
	OTReading := GetOldTestamentDailyRange(t.YearDay(), []string{})
	stmt := fmt.Sprintf("select book, chapter, verse, text from kjv where ordinal_verse between %d and %d", OTReading.StartOrdinalVerse, OTReading.EndOrdinalVerse)
	fmt.Println(stmt)
	rows, err := app.Database.Query(stmt)

	if err != nil {
		log.Fatalf("Failed to get verses for OT Reading")
	}

	for rows.Next() {
		v := Verse{}
		rows.Scan(&v.Book, &v.Chapter, &v.Verse, &v.Text)
		// fmt.Printf("%#v\n", v)
		versesFromOT = append(versesFromOT, v)
	}

	// TODO: Render HTML response , just JSON for now cause time
	jsonizeResponse(versesFromOT, w)

}

func (app *App) GetDailyNewTestament(w http.ResponseWriter, r *http.Request) {
	fmt.Println("New Testament Daily Range")
	versesFromNT := []Verse{}

	t := time.Now()
	NTReading := GetNewTestamentDailyRange(t.YearDay())
	stmt := fmt.Sprintf("select book, chapter, verse, text from kjv where ordinal_verse between %d and %d", NTReading.StartOrdinalVerse, NTReading.EndOrdinalVerse)
	fmt.Println(stmt)
	rows, err := app.Database.Query(stmt)

	if err != nil {
		log.Fatalf("Failed to get verses for NT Reading")
	}

	for rows.Next() {
		v := Verse{}
		rows.Scan(&v.Book, &v.Chapter, &v.Verse, &v.Text)
		// fmt.Printf("%#v\n", v)
		versesFromNT = append(versesFromNT, v)
	}

	// TODO: Render HTML response , just JSON for now cause time
	jsonizeResponse(versesFromNT, w)

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

	stmt := ""

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

		// Create the sqlverseRange
		sqlVerseRange := ""
		for i := verses.StartVerse; i < verses.EndVerse; i++ {
			sqlVerseRange = sqlVerseRange + strconv.Itoa(i) + ","
		}
		sqlVerseRange = sqlVerseRange + strconv.Itoa(verses.EndVerse)

		log.Printf("sql verse range: %s", sqlVerseRange)

		stmt = fmt.Sprintf("select verse, text from kjv where book=\"%s\" and chapter=%s and verse in (%s)\n",
			bookName,
			strconv.Itoa(rChapter),
			sqlVerseRange)

		log.Printf("Multi verse sql query: %s", stmt)

		// create HTML Title
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

		// Query the database
		stmt = fmt.Sprintf("select verse, text from kjv where book=\"%s\" and chapter=%s and verse=%s",
			bookName,
			strconv.Itoa(rChapter),
			strconv.Itoa(rVerse),
		)

		log.Printf("Single verse sql query: %s\n", stmt)

		// create HTML Title
		verses.HTMLTitle = fmt.Sprintf("%s %s:%s",
			bookName,
			strconv.Itoa(rChapter),
			requestVars["verse"],
		)
	}

	rows, err := app.Database.Query(stmt)
	if err != nil {
		http.Error(w, "Could not query DB", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var verseNum int
	var text string

	for rows.Next() {
		rows.Scan(&verseNum, &text)
		verses.Verses = append(verses.Verses, map[int]string{verseNum: text})
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

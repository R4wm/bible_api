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
	kjv "github.com/r4wm/bible_api"
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
	app.Router.HandleFunc("/bible/list_books", app.listBooks) // why do i have to be explicit about the post slash here..

	s := app.Router.PathPrefix("/bible").Subrouter()
	s.HandleFunc("/{book}", app.getBook)
	s.HandleFunc("/{book}/{chapter}", app.getChapter)
	s.HandleFunc("/{book}/{chapter}/{verse}", app.getVerse)

}

func (app *App) listBooks(w http.ResponseWriter, r *http.Request) {
	r.URL.Path = strings.TrimSuffix(r.URL.Path, "/")
	fmt.Println("removed slas")
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

	booksStruct := struct {
		Books []string
	}{
		Books: books,
	}

	// Return json response if requested
	jsonizeResponse(booksStruct, w, r)
	return
}

// TODO this should return all verses for the book in json
func (app *App) getBook(w http.ResponseWriter, r *http.Request) {
	r.URL.Path = strings.TrimSuffix(r.URL.Path, "/")

	vars := mux.Vars(r)
	var chapters struct {
		BookName     string
		ChapterCount int
	}
	chapters.BookName = strings.ToUpper(vars["book"])
	chapters.ChapterCount = BookChapterLimit[chapters.BookName]

	jsonizeResponse(chapters, w, r)
	return
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
		Verses       []Verse
		SearchString string
		Count        map[string]int
		GraphCount   string // json array of ints
	}

	graphBookCounter := [66]int{}
	var defaultSearchLimit = "100000"

	// Handle text query
	searchText, ok := r.URL.Query()["q"]
	fmt.Printf("%v\n", searchText)
	if !ok || len(searchText) < 1 {
		w.Write([]byte("Ye ask, and receive not, because ye ask amiss, that ye may consume it upon your lusts."))
		return
	}

	// Handle limit size
	searchLimit, ok := r.URL.Query()["n"]
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

	matches.SearchString = searchText[0]

	rows, err := app.Database.Query("select book, chapter, verse, text, ordinal_book from kjv where text like ? limit ?", "%"+searchText[0]+"%", limit)
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

	tmpl, err := template.New("results").Funcs(funcs).Parse(searchResultTemplate)
	if err != nil {
		fmt.Println("Failed to parse template..")
		return
	}

	graphBytes, err := json.Marshal(graphBookCounter)
	if err != nil {
		log.Fatal("bad search json")
	}

	matches.GraphCount = string(graphBytes)

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

	jsonizeResponse(result, w, r)
	return
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
		verses    = []Verse{}
		vars      = mux.Vars(r)
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
	BookName := vars["book"]

	chapter, err := strconv.Atoi(vars["chapter"])
	if err != nil {
		w.WriteHeader(http.StatusNotAcceptable)
		w.Write([]byte(fmt.Sprintf("%s is not a proper chapter", vars["chapter"])))
		return
	}
	// check the chapter is not > last chapter for book
	if chapter > BookChapterLimit[BookName] {
		w.WriteHeader(http.StatusNotAcceptable)
		msg := fmt.Sprintf("Chapter %d is out of bounds, last chapter of %s is %d\n", chapter, BookName, BookChapterLimit[BookName])
		w.Write([]byte(msg))
		return
	}
	stmt := fmt.Sprintf("select verse, text from kjv where book='%s' and chapter=%v", BookName, chapter)
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
		verses = append(verses, Verse{vars["book"], chapter, verse, text})
	}
	jsonizeResponse(verses, w, r)
}

func (app *App) GetDailyProverbs(w http.ResponseWriter, r *http.Request) {

	versesFromProverbs := []Verse{}

	proverbsReading := GetProverbsDailyRange(GetDaysInMonth(), time.Now().Day())
	fmt.Printf("%#v\n", proverbsReading)

	stmt := fmt.Sprintf("select book, chapter, verse, text from kjv where ordinal_verse between %d and %d", proverbsReading.StartOrdinalVerse, proverbsReading.EndOrdinalVerse)
	fmt.Println(stmt)

	rows, err := app.Database.Query(stmt)
	if err != nil {
		log.Fatalf("Failed to query DAtabase")
	}

	for rows.Next() {
		v := Verse{}
		rows.Scan(&v.Book, &v.Chapter, &v.Verse, &v.Text)
		// fmt.Printf("%#v\n", v)
		versesFromProverbs = append(versesFromProverbs, v)
	}

	// TODO: Render HTML response , just JSON for now cause time
	jsonizeResponse(versesFromProverbs, w, r)
	return
}

func (app *App) GetDailyPsalms(w http.ResponseWriter, r *http.Request) {

	versesFromPsalms := []Verse{}

	proverbsReading := GetPsalmsDailyRange(GetDaysInMonth(), time.Now().Day())
	fmt.Printf("%#v\n", proverbsReading)

	stmt := fmt.Sprintf("select book, chapter, verse, text from kjv where ordinal_verse between %d and %d", proverbsReading.StartOrdinalVerse, proverbsReading.EndOrdinalVerse)
	fmt.Println(stmt)

	rows, err := app.Database.Query(stmt)
	if err != nil {
		log.Fatalf("Failed to query DAtabase")
	}

	for rows.Next() {
		v := Verse{}
		rows.Scan(&v.Book, &v.Chapter, &v.Verse, &v.Text)
		// fmt.Printf("%#v\n", v)
		versesFromPsalms = append(versesFromPsalms, v)
	}

	// TODO: Render HTML response , just JSON for now cause time
	jsonizeResponse(versesFromPsalms, w, r)
	return
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
	jsonizeResponse(versesFromOT, w, r)
	return

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

	verses.BookName = requestVars["book"]

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
			http.Error(w, fmt.Sprintf("Verse range end is not valid: %s", verseRange[1]), http.StatusBadRequest)
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
			requestVars["book"],
			strconv.Itoa(rChapter),
			sqlVerseRange)

		log.Printf("Multi verse sql query: %s", stmt)

		// create HTML Title
		verses.HTMLTitle = fmt.Sprintf("%s %s:%s-%s",
			requestVars["book"],
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
			requestVars["book"],
			strconv.Itoa(rChapter),
			strconv.Itoa(rVerse),
		)

		log.Printf("Single verse sql query: %s\n", stmt)

		// create HTML Title
		verses.HTMLTitle = fmt.Sprintf("%s %s:%s",
			requestVars["book"],
			strconv.Itoa(rChapter),
			requestVars["verse"],
		)
	}

	rows, err := app.Database.Query(stmt)
	if err != nil {
		http.Error(w, fmt.Sprintf("Could not query DB"), http.StatusInternalServerError)
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
		jsonizeResponse(verses, w, r)
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

package kjvapi

import (
	"bufio"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

// Verse the complete verse context
type Verse struct {
	IsNumberedBook bool
	Book           string
	Chapter        int
	Verse          int
	Text           string
	Testament      string
	OrdinalVerse   int
	OrdinalBook    int
}

//ParseChapterVerse extract chapter and verse from x:x format
func ParseChapterVerse(colonJoined string) (int, int) {
	fmt.Printf("colonJoined: %v\n", colonJoined)

	splitChapterVerse := strings.Split(colonJoined, ":")

	chapter, err := strconv.Atoi(splitChapterVerse[0])
	if err != nil {
		panic(err)
	}

	verseNum, err := strconv.Atoi(splitChapterVerse[1])
	if err != nil {
		panic(err)
	}

	return chapter, verseNum

}

// IsNumberedBook determines if this is numbered book like 1John or 2Timothy.
func IsNumberedBook(firstPart string) bool {
	// firstPart is the very first element in the parsed string.
	if _, err := strconv.Atoi(firstPart); err == nil {
		return true
	}
	return false
}

//PrepareDB Inserts verse context into database. Note old db WILL be deleted.
func PrepareDB(verse <-chan Verse, dbPath string) {

	// Delete old existing database
	if _, err := os.Stat(dbPath); !os.IsNotExist(err) {
		if err := os.Remove(dbPath); err != nil {
			log.Fatalf("Could not remove old database: %s", dbPath)
		}
	}

	//Create new database
	database, _ := sql.Open("sqlite3", dbPath) // "data/kjv.sqlite3.db")
	defer database.Close()

	//Prep new database
	statement, _ := database.Prepare("create table if not exists kjv(book string not null, chapter int, verse int, text string, ordinal_verse int, ordinal_book int, testament string)")
	statement.Exec()

	sqlInsertStr := `INSERT OR REPLACE INTO kjv(book, chapter, verse, text, ordinal_verse, ordinal_book, testament) values(?, ?, ?, ?, ?, ?, ?)`
	stmt, err := database.Prepare(sqlInsertStr)
	if err != nil {
		panic(err)
	}

	//Populate, put into database as they come
	defer stmt.Close()
	for v := range verse {
		stmt.Exec(v.Book, v.Chapter, v.Verse, v.Text, v.OrdinalVerse, v.OrdinalBook, v.Testament)
	}
}

//CreateKJVDB pulls down KJV raw text file, parses and creates database
func CreateKJVDB(dbpath string) string {
	fmt.Println("Starting sqlite3 db creation. ")

	url := "https://raw.githubusercontent.com/R4wm/bible/master/data/bible.txt"
	dbInsert := make(chan Verse)

	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	go PrepareDB(dbInsert, dbpath)

	scanner := bufio.NewScanner(resp.Body)

	verseCount := 0
	bookCount := 0
	bookNameState := ""
	bookTestament := "Old"
	for scanner.Scan() {
		verseCount += 1
		verse := Verse{}
		brokenString := strings.Fields(scanner.Text())
		fmt.Println("broken: ", brokenString)

		// First book of the New Testament
		if brokenString[0] == "Matthew" {
			bookTestament = "New"
		}

		if brokenString[0] != bookNameState {
			bookNameState = brokenString[0]
			bookCount += 1
		}

		if brokenString[0] == "Song" {
			//This is Song of Solomon book, special case where book name has multiple words
			verse.Book = fmt.Sprintf("%s %s %s",
				strings.ToUpper(brokenString[0]), // SONG
				strings.ToUpper(brokenString[1]), // OF
				strings.ToUpper(brokenString[2])) // SOLOMON
			verse.Chapter, verse.Verse = ParseChapterVerse(brokenString[3])
			verse.Text = strings.Join(brokenString[4:], " ")

		} else if IsNumberedBook(brokenString[0]) {
			verse.Book = strings.ToUpper(brokenString[0] + brokenString[1])
			verse.Chapter, verse.Verse = ParseChapterVerse(brokenString[2])
			verse.Text = strings.Join(brokenString[3:], " ")

		} else {
			verse.Book = strings.ToUpper(brokenString[0])
			verse.Chapter, verse.Verse = ParseChapterVerse(brokenString[1])
			verse.Text = strings.Join(brokenString[2:], " ")
		}

		verse.OrdinalBook = bookCount
		verse.OrdinalVerse = verseCount
		verse.Testament = strings.ToUpper(bookTestament)
		fmt.Printf("verse: %v\n", verse)

		dbInsert <- verse
	}

	close(dbInsert)

	return "dbpath"
}

package main

import (
	"html/template"
	"os"
)

type book struct {
	Name      string
	Testament string
}

type Books struct {
	List []book
}

func main() {

	var KJV Books

	kjvBooks := []string{"GENESIS", "EXODUS", "LEVITICUS", "NUMBERS", "DEUTERONOMY", "JOSHUA", "JUDGES", "RUTH", "1SAMUEL", "2SAMUEL", "1KINGS", "2KINGS", "1CHRONICLES", "2CHRONICLES", "EZRA", "NEHEMIAH", "ESTHER", "JOB", "PSALMS", "PROVERBS", "ECCLESIASTES", "SONG OF SOLOMON", "ISAIAH", "JEREMIAH", "LAMENTATIONS", "EZEKIEL", "DANIEL", "HOSEA", "JOEL", "AMOS", "OBADIAH", "JONAH", "MICAH", "NAHUM", "HABAKKUK", "ZEPHANIAH", "HAGGAI", "ZECHARIAH", "MALACHI", "MATTHEW", "MARK", "LUKE", "JOHN", "ACTS", "ROMANS", "1CORINTHIANS", "2CORINTHIANS", "GALATIANS", "EPHESIANS", "PHILIPPIANS", "COLOSSIANS", "1THESSALONIANS", "2THESSALONIANS", "1TIMOTHY", "2TIMOTHY", "TITUS", "PHILEMON", "HEBREWS", "JAMES", "1PETER", "2PETER", "1JOHN", "2JOHN", "3JOHN", "JUDE", "REVELATION"}

	for i := 0; i < len(kjvBooks); i++ {
		var entry book

		if i >= 39 {
			entry = book{Name: kjvBooks[i], Testament: "new"}
		} else {
			entry = book{Name: kjvBooks[i], Testament: "old"}
		}

		KJV.List = append(KJV.List, entry)
	}

	// Make the template
	paths := []string{"another-template.html"}

	t := template.Must(template.New("another-template.html").ParseFiles(paths...))

	err := t.Execute(os.Stdout, KJV)

	if err != nil {
		panic(err)
	}
}

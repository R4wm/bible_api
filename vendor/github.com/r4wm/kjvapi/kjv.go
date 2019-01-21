package kjvapi

//KJVVerse simple container for verses
type KJVVerse struct {
	Verse int    `json:"verse"`
	Text  string `json:"text"`
}

//KJVChapter simple container for chapters
type KJVChapter struct {
	Chapter int `json:"chapter"`
	Verses  []KJVVerse
}

//KJVBook simpla container for book from Bible
type KJVBook struct {
	Book     string `json:"book"`
	Chapters []KJVChapter
}

// GetChapter compose a bible chapter with verses
func GetChapter(book string, chapter int, verses []KJVVerse) *KJVBook {
	return &KJVBook{
		Book: book,
		Chapters: []KJVChapter{
			KJVChapter{
				Chapter: chapter,
				Verses:  verses}},
	}
}

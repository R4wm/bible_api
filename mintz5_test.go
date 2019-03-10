package mintz5

import (
	"fmt"
	"testing"
)

func TestGetRandomColor(t *testing.T) {
	color := GetRandomColor()
	colorType := fmt.Sprintf("%T", color)

	if colorType != "string" {
		t.Errorf("Expected string got %v\n", colorType)
	}
}

func TestGetBookMaxChapter(t *testing.T) {
	//////////////////////////
	// Positive Test        //
	//////////////////////////
	book := "ROMANS"
	result := GetBookMaxChapter(book)

	if result != 16 {
		t.Errorf("Expected 16 chapters for Romans, got something else: %v\n", result)
	}
	//////////////////////////
	// Negative Test        //
	//////////////////////////
	nonExistentBook := "IDONTEXIST"
	failResult := GetBookMaxChapter(nonExistentBook)

	if failResult != 0 {
		t.Errorf("Expected 0 for fake book: %s, got %v\n", nonExistentBook, failResult)
	}
}

package kjv

import (
	"testing"
)

func TestGetOldTestamentDailyRangeStartFistDay(t *testing.T) {
	expectedStart := 1
	expectedEnd := 64
	result := GetOldTestamentDailyRange(1, []string{})
	if result.StartOrdinalVerse != expectedStart {
		t.Errorf("Expected %d, got %d", expectedStart, result.StartOrdinalVerse)
	}

	if result.EndOrdinalVerse != expectedEnd {
		t.Errorf("Last verse expected: %d, got %d\n", expectedStart, expectedEnd)
	}
}

func TestGetOldTestamentDailyRangeVerseCount(t *testing.T) {
	expectedCount := 63 // 1954 - 1891
	result := GetOldTestamentDailyRange(31, []string{})
	if result.TotalVerseCount != expectedCount {
		t.Errorf("Expected %d got %d", expectedCount, result.TotalVerseCount)
	}
}

func TestGetOldTestamentDailyRangeStartSecondDay(t *testing.T) {
	expectedStart := 64
	result := GetOldTestamentDailyRange(2, []string{})
	if result.StartOrdinalVerse != expectedStart {
		t.Errorf("Expected %d, got %d", expectedStart, result.StartOrdinalVerse)
	}
}

func TestGetOldTestamentDailyRangeEnd(t *testing.T) {
	expectedEnd := 1954
	result := GetOldTestamentDailyRange(31, []string{})
	if result.EndOrdinalVerse != expectedEnd {
		t.Errorf("Expected %d got %d", expectedEnd, result.EndOrdinalVerse)
	}
}

func TestGetOldTestamentDailyRangeLastDay(t *testing.T) {
	expectedLastOTVerse := 23145
	result := GetOldTestamentDailyRange(365, []string{})
	if result.EndOrdinalVerse != expectedLastOTVerse {
		t.Errorf("Last OT ordinal verse is wrong, should be %d got %d",
			expectedLastOTVerse, result.EndOrdinalVerse)
	}
}

func TestOTExceptProverbs(t *testing.T) {
	expectedLastOTVerse := 23145
	result := GetOldTestamentDailyRange(365, []string{"proverbs"})
	if result.EndOrdinalVerse != expectedLastOTVerse {
		t.Errorf("Last OT ordinal verse is wrong, should be %d got %d",
			expectedLastOTVerse, result.EndOrdinalVerse)
	}

}

// func TestProverbsExclusionFromOTRange(t *testing.T) {
// 	var proverbsEnd int = PsalmsOrdinalStart - 1
// 	for i := 1; i < 365; i++ {
// 		result := GetOldTestamentDailyRange(i, []string{"proverbs"})
// 		if !(result.StartOrdinalVerse >= ProverbsOrdinalStart || result.StartOrdinalVerse <= proverbsEnd) {
// 			t.Errorf("Proverbs Start verse found when excluded: %#v\n", result)
// 		}

// 	}
// }

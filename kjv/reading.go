package kjv

import (
	"math"
	"time"
)

// sqlite> select count(text) from kjv where book="PSALMS";
// 2461
// sqlite> select count(text) from kjv where book="PROVERBS";
// 915
// sqlite> select count(text) from kjv where testament="OLD";
// 23145
// sqlite> select count(text) from kjv where testament="NEW";
// 7956
// sqlite>

////////////////////////////////////////////////////////////////////////
// TODO: Most of these functions could be reusable, consolidate later //
////////////////////////////////////////////////////////////////////////
const (
	DaysInYear             = 365
	VerseCountOldTestament = 23145
	VerseCountNewTestament = 7956
	VerseCountPsalms       = 2461
	VerseCountProverbs     = 915
	PsalmsOrdinalStart     = 13941 // This is where Psalms starts
	ProverbsOrdinalStart   = 16402 // This is where Proverbs starts
	PsalmsOrdinalEnd       = ProverbsOrdinalStart - 1
	ProverbsOrdinalEnd     = 17316
	FirstOrdinalVerseOT    = 1
	FirstOrdinalVerseNT    = 23146
	TotalVersesInBible     = 31101
)

type ReadingSchedule struct {
	StartOrdinalVerse int
	EndOrdinalVerse   int
	TotalVerseCount   int
}

func GetDaysInMonth() int {
	t := time.Now()
	return time.Date(t.Year(), t.Month()+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

func GetProverbsDailyRange(daysInMonth, currentDay int) ReadingSchedule {
	// fmt.Println(kjv.GetProverbsDailyRange(kjv.GetDaysInMonth(), time.Now().Day()))

	verseCountPerDay := int(math.Ceil(float64(VerseCountProverbs / daysInMonth)))

	// Set beginning ordinal verse number for today
	startReadingAt := ProverbsOrdinalStart
	for i := 1; i < currentDay; i++ {
		startReadingAt += verseCountPerDay
	}

	// Because we round off, we have to detemine if this is the last day of the month
	isLastDay := daysInMonth == currentDay

	if isLastDay {
		return ReadingSchedule{
			StartOrdinalVerse: startReadingAt,
			EndOrdinalVerse:   ProverbsOrdinalEnd,
			TotalVerseCount:   ProverbsOrdinalEnd - startReadingAt,
		}
	}

	// isNotLastDay
	return ReadingSchedule{
		StartOrdinalVerse: startReadingAt,
		EndOrdinalVerse:   startReadingAt + verseCountPerDay,
		TotalVerseCount:   (startReadingAt + verseCountPerDay) - startReadingAt,
	}
}

func GetPsalmsDailyRange(daysInMonth, currentDay int) ReadingSchedule {
	// fmt.Println(kjv.GetPsalmsDailyRange(kjv.GetDaysInMonth(), time.Now().Day()))

	verseCountPerDay := int(math.Ceil(float64(VerseCountPsalms / daysInMonth)))

	// Set beginning ordinal verse number for today
	startReadingAt := PsalmsOrdinalStart
	for i := 1; i < currentDay; i++ {
		startReadingAt += verseCountPerDay
	}

	// Because we round off, we have to detemine if this is the last day of the month
	isLastDay := daysInMonth == currentDay

	if isLastDay {
		return ReadingSchedule{
			StartOrdinalVerse: startReadingAt,
			EndOrdinalVerse:   PsalmsOrdinalEnd,
			TotalVerseCount:   PsalmsOrdinalEnd - startReadingAt,
		}
	}

	// isNotLastDay
	return ReadingSchedule{
		StartOrdinalVerse: startReadingAt,
		EndOrdinalVerse:   startReadingAt + verseCountPerDay,
		TotalVerseCount:   (startReadingAt + verseCountPerDay) - startReadingAt,
	}
}

func GetOldTestamentDailyRange(currentDayofYear int, excludedBooks []string) ReadingSchedule {
	verseCountPerDay := 63
	// Set beginning ordinal verse number for today
	startReadingAt := 1
	for i := 1; i < currentDayofYear; i++ {
		startReadingAt += verseCountPerDay
	}

	// Last Day
	if DaysInYear == currentDayofYear {
		lastOrdinalVerseOT := FirstOrdinalVerseNT - 1
		return ReadingSchedule{
			StartOrdinalVerse: startReadingAt,
			EndOrdinalVerse:   lastOrdinalVerseOT,
			TotalVerseCount:   lastOrdinalVerseOT - startReadingAt,
		}
	}

	// isNotLastDay or FirstDay
	return ReadingSchedule{
		StartOrdinalVerse: startReadingAt,
		EndOrdinalVerse:   startReadingAt + verseCountPerDay,
		TotalVerseCount:   (startReadingAt + verseCountPerDay) - startReadingAt,
	}
}

func GetNewTestamentDailyRange(currentDayofYear int) ReadingSchedule {

	// verseCountPerDay := int(math.Ceil(float64(VerseCountNewTestament / DaysInYear)))

	verseCountPerDay := 23
	startReadingAt := FirstOrdinalVerseNT
	for i := 1; i < currentDayofYear; i++ {
		startReadingAt += verseCountPerDay
	}

	isLastDay := DaysInYear == currentDayofYear
	if isLastDay {
		return ReadingSchedule{
			StartOrdinalVerse: startReadingAt,
			EndOrdinalVerse:   FirstOrdinalVerseNT - 1,
			TotalVerseCount:   (FirstOrdinalVerseNT - 1) - startReadingAt,
		}
	}

	// isNotLastDay
	return ReadingSchedule{
		StartOrdinalVerse: startReadingAt,
		EndOrdinalVerse:   startReadingAt + verseCountPerDay,
		TotalVerseCount:   (startReadingAt + verseCountPerDay) - startReadingAt,
	}
}

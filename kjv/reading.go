package kjv

import (
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

package kjv

import (
	"fmt"
	"testing"
)

func TestStuff(t *testing.T) {
	otResult, ntResult := getReadingRanges(1)
	fmt.Println(otResult, ntResult)
	fmt.Printf("%#v\n", otResult)
	fmt.Printf("%T %T\n", otResult, ntResult)
	fmt.Println("ok")

	dayOneVerse := 1
	if otResult.StartOrdinalVerse != dayOneVerse {
		t.Fatalf("Expected %d for starting OT Verse, got %d\n",
			dayOneVerse,
			otResult.StartOrdinalVerse)
	}

	if otResult.EndOrdinalVerse != 23 {
		t.Fatalf("Expected %d for starting OT Verse, got %d\n",
			23,
			otResult.EndOrdinalVerse)
	}

	if ntResult.StartOrdinalVerse != dayOneVerse {
		t.Fatalf("Expected %d for starting OT Verse, got %d\n",
			dayOneVerse,
			ntResult.StartOrdinalVerse)
	}

	if ntResult.EndOrdinalVerse != 64 {
		t.Fatalf("Expected %d for starting OT Verse, got %d\n",
			64,
			ntResult.EndOrdinalVerse)
	}

}

package kjv

import "testing"

func TestGetRandomColorReturnsKnownColor(t *testing.T) {
	if len(Colors) == 0 {
		t.Fatal("Colors slice is empty")
	}

	c := GetRandomColor()
	if c == "" {
		t.Fatal("GetRandomColor returned empty string")
	}

	found := false
	for _, color := range Colors {
		if c == color {
			found = true
			break
		}
	}

	if !found {
		t.Fatalf("GetRandomColor returned unknown color: %q", c)
	}
}

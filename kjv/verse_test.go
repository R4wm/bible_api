package kjv

import "testing"

func TestRemoveItalicMarkers(t *testing.T) {
	v := Verse{Text: "[In] the [beginning]"}
	v.RemoveItalicMarkers()
	if v.Text != "In the beginning" {
		t.Fatalf("unexpected text after RemoveItalicMarkers: %q", v.Text)
	}
}

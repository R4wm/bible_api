package kjv

import "testing"

func TestRemoveItalicMarkers(t *testing.T) {
	v := Verse{Text: "[In] the [beginning]"}
	v.RemoveItalicMarkers()
	if v.Text != "In the beginning" {
		t.Fatalf("unexpected text after RemoveItalicMarkers: %q", v.Text)
	}
}

func TestParseSuggestSearchResponseIncludesReference(t *testing.T) {
	body := []byte(`{
		"took": 3,
		"hits": {
			"hits": [
				{
					"_score": 12.5,
					"_source": {
						"text": "And we know that all things work together for good to them that love God...",
						"book": "ROMANS",
						"chapter": 8,
						"verse": 28
					}
				}
			]
		}
	}`)

	suggestions, took, err := parseSuggestSearchResponse(body)
	if err != nil {
		t.Fatalf("parseSuggestSearchResponse returned error: %v", err)
	}
	if took != 3 {
		t.Fatalf("unexpected took value: %d", took)
	}
	if len(suggestions) != 1 {
		t.Fatalf("expected 1 suggestion, got %d", len(suggestions))
	}

	got := suggestions[0]
	if got["text"] == "" || got["score"] != 12.5 || got["book"] != "ROMANS" || got["chapter"] != 8 || got["verse"] != 28 {
		t.Fatalf("suggestion did not include expected reference metadata: %#v", got)
	}
}

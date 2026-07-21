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

func TestBuildSearchBodyUsesCanonicalVerseOrder(t *testing.T) {
	body := buildSearchBody("grace", 50, 0, map[string]string{})
	sortFields, ok := body["sort"].([]interface{})
	if !ok || len(sortFields) != 1 {
		t.Fatalf("expected one sort field, got %#v", body["sort"])
	}
	field, ok := sortFields[0].(map[string]interface{})
	if !ok {
		t.Fatalf("unexpected sort field: %#v", sortFields[0])
	}
	if _, ok := field["ordinal_verse"]; !ok {
		t.Fatalf("expected ordinal_verse sort, got %#v", field)
	}
}

func TestBuildSuggestSearchBodyUsesCanonicalVerseOrder(t *testing.T) {
	body := buildSuggestSearchBody("thy word", 20, 0)
	sortFields, ok := body["sort"].([]interface{})
	if !ok || len(sortFields) != 1 {
		t.Fatalf("expected one sort field, got %#v", body["sort"])
	}
	field, ok := sortFields[0].(map[string]interface{})
	if !ok {
		t.Fatalf("unexpected sort field: %#v", sortFields[0])
	}
	if _, ok := field["ordinal_verse"]; !ok {
		t.Fatalf("expected ordinal_verse sort, got %#v", field)
	}
}

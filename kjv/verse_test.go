package kjv

import (
	"net/http"
	"testing"
)

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
	body := buildSearchBody("grace", 50, 0, map[string]string{}, searchOptions{Match: searchMatchAny})
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

func TestBuildSearchBodyMatchModes(t *testing.T) {
	tests := []struct {
		name      string
		options   searchOptions
		queryKind string
		field     string
		operator  string
	}{
		{"any words", searchOptions{Match: searchMatchAny}, "match", "text", ""},
		{"all words", searchOptions{Match: searchMatchAll}, "match", "text", "and"},
		{"exact phrase", searchOptions{Match: searchMatchPhrase}, "match_phrase", "text", ""},
		{"case-sensitive all words", searchOptions{Match: searchMatchAll, CaseSensitive: true}, "match", "text.case_sensitive", "and"},
		{"case-sensitive phrase", searchOptions{Match: searchMatchPhrase, CaseSensitive: true}, "match_phrase", "text.case_sensitive", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := buildSearchBody("Thy Word", 50, 0, map[string]string{}, tt.options)
			query := body["query"].(map[string]interface{})["bool"].(map[string]interface{})["must"].([]interface{})[0].(map[string]interface{})
			clause, ok := query[tt.queryKind].(map[string]interface{})
			if !ok {
				t.Fatalf("expected %s clause, got %#v", tt.queryKind, query)
			}
			value, ok := clause[tt.field].(map[string]interface{})
			if !ok || value["query"] != "Thy Word" {
				t.Fatalf("expected query on %s, got %#v", tt.field, clause)
			}
			if got, _ := value["operator"].(string); got != tt.operator {
				t.Fatalf("operator = %q, want %q", got, tt.operator)
			}
		})
	}
}

func TestParseSearchOptions(t *testing.T) {
	req, _ := http.NewRequest("GET", "/bible/v2/search?q=Thy+Word&match=phrase&case_sensitive=true", nil)
	got, err := parseSearchOptions(req)
	if err != nil {
		t.Fatalf("parseSearchOptions: %v", err)
	}
	if got.Match != searchMatchPhrase || !got.CaseSensitive {
		t.Fatalf("unexpected options: %#v", got)
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

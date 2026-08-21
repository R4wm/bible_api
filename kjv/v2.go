package kjv

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

const (
	defaultOpenSearchURL   = "http://localhost:9200"
	defaultOpenSearchIndex = "kjv_v2"
	defaultSynonymsSet     = "kjv_synonyms"
)

// searchMatchMode defines how multiple terms in q are combined.
// The zero/default value deliberately preserves the existing broad search.
type searchMatchMode string

const (
	searchMatchAny    searchMatchMode = "any"
	searchMatchAll    searchMatchMode = "all"
	searchMatchPhrase searchMatchMode = "phrase"
)

type searchOptions struct {
	Match         searchMatchMode
	CaseSensitive bool
}

type v2SearchResult struct {
	Book      string   `json:"book"`
	Chapter   int      `json:"chapter"`
	Verse     int      `json:"verse"`
	Text      string   `json:"text"`
	Highlight []string `json:"highlight,omitempty"`
}

type v2SearchResponse struct {
	Status string                 `json:"status"`
	Meta   map[string]interface{} `json:"meta"`
	Data   map[string]interface{} `json:"data"`
}

func (app *App) SetupV2Routes() {
	v2 := app.Router.PathPrefix("/bible/v2").Subrouter()
	v2.HandleFunc("/search", app.searchV2).Methods("GET")
	v2.HandleFunc("/suggest", app.suggestV2).Methods("GET")

	syn := v2.PathPrefix("/synonyms").Subrouter()
	syn.Use(app.jwtMiddleware("synonyms:write"))
	syn.HandleFunc("/{set}", app.synonymsV2).Methods("PUT", "POST", "DELETE")
}

func (app *App) InitOpenSearch() {
	if app.OpenSearchURL == "" {
		return
	}
	if app.OpenSearchHTTP == nil {
		app.OpenSearchHTTP = &http.Client{Timeout: 10 * time.Second}
	}

	synSet := getEnvDefault("OPENSEARCH_SYNONYMS_SET", defaultSynonymsSet)
	_ = app.ensureSynonymsSet(synSet)

	if app.OpenSearchIndex == "" {
		app.OpenSearchIndex = defaultOpenSearchIndex
	}

	exists, err := app.openSearchIndexExists(app.OpenSearchIndex)
	if err != nil {
		fmt.Printf("OpenSearch index check failed: %v\n", err)
		return
	}
	if exists {
		fmt.Printf("OpenSearch index '%s' already exists. If /bible/suggest returns poor results, reindex with: curl -X DELETE http://localhost:9200/%s && python3 scripts/index_kjv_to_opensearch.py\n", app.OpenSearchIndex, app.OpenSearchIndex)
		return
	}

	mapping, err := loadOpenSearchMapping()
	if err != nil {
		fmt.Printf("OpenSearch mapping load failed: %v\n", err)
		return
	}
	if err := app.createOpenSearchIndex(app.OpenSearchIndex, mapping); err != nil {
		fmt.Printf("OpenSearch index create failed: %v\n", err)
	}
}

func (app *App) searchV2(w http.ResponseWriter, r *http.Request) {
	started := time.Now()
	result := "error"
	defer func() {
		app.Analytics.RecordSearch("v2", result, time.Since(started))
	}()

	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if query == "" {
		result = "invalid"
		jsonError(w, http.StatusBadRequest, "q is required")
		return
	}
	options, err := parseSearchOptions(r)
	if err != nil {
		result = "invalid"
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	// Validate the request before checking a downstream dependency. Clients
	// should receive a stable 400 for invalid options even while OpenSearch is
	// unavailable.
	if err := app.ensureOpenSearchReady(); err != nil {
		jsonError(w, http.StatusServiceUnavailable, err.Error())
		return
	}

	size := parseIntDefault(r.URL.Query().Get("n"), 50)
	if size > 1000 {
		size = 1000
	}
	from := parseIntDefault(r.URL.Query().Get("from"), 0)
	if from < 0 {
		from = 0
	}

	filters := map[string]string{
		"book":      strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("book"))),
		"testament": strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("testament"))),
		"chapter":   strings.TrimSpace(r.URL.Query().Get("chapter")),
		"verse":     strings.TrimSpace(r.URL.Query().Get("verse")),
	}

	searchBody := buildSearchBody(query, size, from, filters, options)
	bodyBytes, _ := json.Marshal(searchBody)

	respBody, status, err := app.doOpenSearchRequest("POST", fmt.Sprintf("/%s/_search", app.OpenSearchIndex), bodyBytes)
	if err != nil {
		jsonError(w, http.StatusBadGateway, err.Error())
		return
	}
	if status >= 400 {
		jsonError(w, http.StatusBadGateway, fmt.Sprintf("OpenSearch error: %s", string(respBody)))
		return
	}

	results, took, total, bookCounts, err := parseSearchResponse(respBody)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "Failed to parse search response")
		return
	}
	result = "ok"

	out := v2SearchResponse{
		Status: "ok",
		Meta: map[string]interface{}{
			"query":          query,
			"count":          total,
			"from":           from,
			"size":           size,
			"took_ms":        took,
			"match":          options.Match,
			"case_sensitive": options.CaseSensitive,
		},
		Data: map[string]interface{}{
			"results":     results,
			"book_counts": bookCounts,
		},
	}
	writeJSON(w, http.StatusOK, out)
}

func (app *App) suggestV2(w http.ResponseWriter, r *http.Request) {
	if err := app.ensureOpenSearchReady(); err != nil {
		jsonError(w, http.StatusServiceUnavailable, err.Error())
		return
	}

	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if query == "" {
		jsonError(w, http.StatusBadRequest, "q is required")
		return
	}

	size := parseIntDefault(r.URL.Query().Get("n"), 10)
	if size > 50 {
		size = 50
	}
	from := parseIntDefault(r.URL.Query().Get("from"), 0)
	if from < 0 {
		from = 0
	}

	body := buildSuggestSearchBody(query, size, from)
	bodyBytes, _ := json.Marshal(body)

	respBody, status, err := app.doOpenSearchRequest("POST", fmt.Sprintf("/%s/_search", app.OpenSearchIndex), bodyBytes)
	if err != nil {
		jsonError(w, http.StatusBadGateway, err.Error())
		return
	}
	if status >= 400 {
		jsonError(w, http.StatusBadGateway, fmt.Sprintf("OpenSearch error: %s", string(respBody)))
		return
	}

	suggestions, took, err := parseSuggestSearchResponse(respBody)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "Failed to parse suggest response")
		return
	}

	out := v2SearchResponse{
		Status: "ok",
		Meta: map[string]interface{}{
			"query":   query,
			"count":   len(suggestions),
			"from":    from,
			"size":    size,
			"took_ms": took,
		},
		Data: map[string]interface{}{
			"suggestions": suggestions,
		},
	}
	writeJSON(w, http.StatusOK, out)
}

func buildSuggestSearchBody(query string, size, from int) map[string]interface{} {
	return map[string]interface{}{
		"query": map[string]interface{}{
			"match_phrase_prefix": map[string]interface{}{
				"text": query,
			},
		},
		"from": from,
		"size": size,
		"sort": []interface{}{
			map[string]interface{}{"ordinal_verse": map[string]string{"order": "asc"}},
		},
		"_source": []string{
			"text",
			"book",
			"chapter",
			"verse",
		},
		"highlight": map[string]interface{}{
			"fields": map[string]interface{}{
				"text": map[string]interface{}{},
			},
		},
	}
}

// suggestCompletion handles GET /bible/suggest?q=... using the OpenSearch
// completion suggester on the text_suggest field. Returns a plain JSON array
// of matching n-gram phrases. Always returns HTTP 200 with valid JSON (empty
// array on any failure) since this is called on every keystroke from the
// search bar across all pages.
func (app *App) suggestCompletion(w http.ResponseWriter, r *http.Request) {
	emptyResult := func() {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("[]"))
	}

	query := strings.TrimSpace(strings.ToLower(r.URL.Query().Get("q")))
	if len(query) < 2 {
		emptyResult()
		return
	}

	if app.OpenSearchURL == "" || app.OpenSearchHTTP == nil {
		emptyResult()
		return
	}

	size := parseIntDefault(r.URL.Query().Get("n"), 20)
	if size > 50 {
		size = 50
	}

	body := map[string]interface{}{
		"suggest": map[string]interface{}{
			"verse-suggest": map[string]interface{}{
				"prefix": query,
				"completion": map[string]interface{}{
					"field":           "text_suggest",
					"size":            size,
					"skip_duplicates": true,
				},
			},
		},
		"_source": false,
	}
	bodyBytes, _ := json.Marshal(body)

	respBody, status, err := app.doOpenSearchRequest("POST", fmt.Sprintf("/%s/_search", app.OpenSearchIndex), bodyBytes)
	if err != nil || status >= 400 {
		emptyResult()
		return
	}

	var parsed osSuggestResp
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		emptyResult()
		return
	}

	suggestions := []string{}
	if entries, ok := parsed.Suggest["verse-suggest"]; ok && len(entries) > 0 {
		for _, opt := range entries[0].Options {
			suggestions = append(suggestions, opt.Text)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(suggestions)
}

func (app *App) synonymsV2(w http.ResponseWriter, r *http.Request) {
	if err := app.ensureOpenSearchReady(); err != nil {
		jsonError(w, http.StatusServiceUnavailable, err.Error())
		return
	}
	set := muxVar(r, "set")
	if set == "" {
		jsonError(w, http.StatusBadRequest, "set is required")
		return
	}

	var payload struct {
		Synonyms []string `json:"synonyms"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	current, _ := app.getSynonymsSet(set)
	switch r.Method {
	case http.MethodPut:
		current = payload.Synonyms
	case http.MethodPost:
		current = mergeSynonyms(current, payload.Synonyms)
	case http.MethodDelete:
		current = removeSynonyms(current, payload.Synonyms)
	default:
		jsonError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	if err := app.putSynonymsSet(set, current); err != nil {
		jsonError(w, http.StatusBadGateway, err.Error())
		return
	}
	_ = app.reloadAnalyzers()

	out := v2SearchResponse{
		Status: "ok",
		Meta: map[string]interface{}{
			"set":   set,
			"count": len(current),
		},
		Data: map[string]interface{}{
			"synonyms": current,
		},
	}
	writeJSON(w, http.StatusOK, out)
}

func (app *App) ensureOpenSearchReady() error {
	if app.OpenSearchURL == "" {
		return errors.New("OpenSearch is not configured")
	}
	if app.OpenSearchHTTP == nil {
		app.OpenSearchHTTP = &http.Client{Timeout: 10 * time.Second}
	}
	if app.OpenSearchIndex == "" {
		app.OpenSearchIndex = defaultOpenSearchIndex
	}
	return nil
}

func parseSearchOptions(r *http.Request) (searchOptions, error) {
	options := searchOptions{Match: searchMatchAny}
	if match := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("match"))); match != "" {
		switch searchMatchMode(match) {
		case searchMatchAny, searchMatchAll, searchMatchPhrase:
			options.Match = searchMatchMode(match)
		default:
			return searchOptions{}, fmt.Errorf("match must be one of: any, all, phrase")
		}
	}
	if raw := strings.TrimSpace(r.URL.Query().Get("case_sensitive")); raw != "" {
		parsed, err := strconv.ParseBool(raw)
		if err != nil {
			return searchOptions{}, fmt.Errorf("case_sensitive must be true or false")
		}
		options.CaseSensitive = parsed
	}
	return options, nil
}

func buildSearchBody(query string, size, from int, filters map[string]string, options searchOptions) map[string]interface{} {
	field := "text"
	if options.CaseSensitive {
		field = "text.case_sensitive"
	}

	var textQuery map[string]interface{}
	if options.Match == searchMatchPhrase {
		textQuery = map[string]interface{}{
			"match_phrase": map[string]interface{}{field: map[string]interface{}{"query": query}},
		}
	} else {
		match := map[string]interface{}{"query": query}
		if options.Match == searchMatchAll {
			match["operator"] = "and"
		}
		textQuery = map[string]interface{}{
			"match": map[string]interface{}{field: match},
		}
	}

	must := []interface{}{textQuery}

	filterClauses := []interface{}{}
	if book := filters["book"]; book != "" {
		filterClauses = append(filterClauses, map[string]interface{}{"term": map[string]interface{}{"book": book}})
	}
	if testament := filters["testament"]; testament != "" {
		filterClauses = append(filterClauses, map[string]interface{}{"term": map[string]interface{}{"testament": testament}})
	}
	if chapter := filters["chapter"]; chapter != "" {
		if v, err := strconv.Atoi(chapter); err == nil {
			filterClauses = append(filterClauses, map[string]interface{}{"term": map[string]interface{}{"chapter": v}})
		}
	}
	if verse := filters["verse"]; verse != "" {
		if v, err := strconv.Atoi(verse); err == nil {
			filterClauses = append(filterClauses, map[string]interface{}{"term": map[string]interface{}{"verse": v}})
		}
	}

	boolQuery := map[string]interface{}{
		"must": must,
	}
	if len(filterClauses) > 0 {
		boolQuery["filter"] = filterClauses
	}

	return map[string]interface{}{
		"query": map[string]interface{}{
			"bool": boolQuery,
		},
		"from": from,
		"size": size,
		// ordinal_verse is assigned during indexing and gives a stable
		// Genesis-to-Revelation order across all result pages.
		"sort": []interface{}{
			map[string]interface{}{"ordinal_verse": map[string]string{"order": "asc"}},
		},
		"track_total_hits": true,
		"highlight": map[string]interface{}{
			"fields": map[string]interface{}{
				"text": map[string]interface{}{},
			},
		},
		"aggs": map[string]interface{}{
			"by_book": map[string]interface{}{
				"terms": map[string]interface{}{
					"field": "book",
					"size":  66,
				},
			},
		},
	}
}

type osSearchResp struct {
	Took int `json:"took"`
	Hits struct {
		Total struct {
			Value int `json:"value"`
		} `json:"total"`
		Hits []struct {
			Source struct {
				Book    string `json:"book"`
				Chapter int    `json:"chapter"`
				Verse   int    `json:"verse"`
				Text    string `json:"text"`
			} `json:"_source"`
			Highlight map[string][]string `json:"highlight"`
		} `json:"hits"`
	} `json:"hits"`
	Aggregations struct {
		ByBook struct {
			Buckets []struct {
				Key      string `json:"key"`
				DocCount int    `json:"doc_count"`
			} `json:"buckets"`
		} `json:"by_book"`
	} `json:"aggregations"`
}

func parseSearchResponse(body []byte) ([]v2SearchResult, int, int, map[string]int, error) {
	var resp osSearchResp
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, 0, 0, nil, err
	}
	results := make([]v2SearchResult, 0, len(resp.Hits.Hits))
	for _, hit := range resp.Hits.Hits {
		item := v2SearchResult{
			Book:    hit.Source.Book,
			Chapter: hit.Source.Chapter,
			Verse:   hit.Source.Verse,
			Text:    hit.Source.Text,
		}
		if hl, ok := hit.Highlight["text"]; ok {
			item.Highlight = hl
		}
		results = append(results, item)
	}
	bookCounts := make(map[string]int, len(resp.Aggregations.ByBook.Buckets))
	for _, bucket := range resp.Aggregations.ByBook.Buckets {
		bookCounts[bucket.Key] = bucket.DocCount
	}
	return results, resp.Took, resp.Hits.Total.Value, bookCounts, nil
}

type osSuggestResp struct {
	Took    int `json:"took"`
	Suggest map[string][]struct {
		Options []struct {
			Text  string  `json:"text"`
			Score float64 `json:"score"`
		} `json:"options"`
	} `json:"suggest"`
}

func parseSuggestResponse(body []byte) ([]map[string]interface{}, int, error) {
	var resp osSuggestResp
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, 0, err
	}
	out := []map[string]interface{}{}
	for _, entries := range resp.Suggest {
		for _, entry := range entries {
			for _, opt := range entry.Options {
				out = append(out, map[string]interface{}{
					"text":  opt.Text,
					"score": opt.Score,
				})
			}
		}
	}
	return out, resp.Took, nil
}

type osSuggestSearchResp struct {
	Took int `json:"took"`
	Hits struct {
		Hits []struct {
			Score  float64 `json:"_score"`
			Source struct {
				Text    string `json:"text"`
				Book    string `json:"book"`
				Chapter int    `json:"chapter"`
				Verse   int    `json:"verse"`
			} `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

func parseSuggestSearchResponse(body []byte) ([]map[string]interface{}, int, error) {
	var resp osSuggestSearchResp
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, 0, err
	}
	out := make([]map[string]interface{}, 0, len(resp.Hits.Hits))
	for _, hit := range resp.Hits.Hits {
		out = append(out, map[string]interface{}{
			"text":    hit.Source.Text,
			"score":   hit.Score,
			"book":    hit.Source.Book,
			"chapter": hit.Source.Chapter,
			"verse":   hit.Source.Verse,
		})
	}
	return out, resp.Took, nil
}

func (app *App) openSearchIndexExists(index string) (bool, error) {
	respBody, status, err := app.doOpenSearchRequest("HEAD", fmt.Sprintf("/%s", index), nil)
	if err != nil {
		return false, err
	}
	_ = respBody
	if status == http.StatusOK {
		return true, nil
	}
	if status == http.StatusNotFound {
		return false, nil
	}
	return false, fmt.Errorf("unexpected status: %d", status)
}

func (app *App) createOpenSearchIndex(index string, mapping []byte) error {
	respBody, status, err := app.doOpenSearchRequest("PUT", fmt.Sprintf("/%s", index), mapping)
	if err != nil {
		return err
	}
	if status >= 400 {
		return fmt.Errorf("create index failed: %s", string(respBody))
	}
	return nil
}

func (app *App) reloadAnalyzers() error {
	respBody, status, err := app.doOpenSearchRequest("POST", fmt.Sprintf("/%s/_reload_search_analyzers", app.OpenSearchIndex), nil)
	if err != nil {
		return err
	}
	if status >= 400 {
		return fmt.Errorf("reload analyzers failed: %s", string(respBody))
	}
	return nil
}

func (app *App) ensureSynonymsSet(set string) error {
	if set == "" {
		return nil
	}
	payload := map[string]interface{}{
		"synonyms": []string{},
	}
	body, _ := json.Marshal(payload)
	_, _, err := app.doOpenSearchRequest("PUT", fmt.Sprintf("/_plugins/_synonyms/%s", set), body)
	return err
}

func (app *App) getSynonymsSet(set string) ([]string, error) {
	respBody, status, err := app.doOpenSearchRequest("GET", fmt.Sprintf("/_plugins/_synonyms/%s", set), nil)
	if err != nil {
		return nil, err
	}
	if status == http.StatusNotFound {
		return []string{}, nil
	}
	if status >= 400 {
		return nil, fmt.Errorf("get synonyms failed: %s", string(respBody))
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return nil, err
	}

	if arr, ok := parsed["synonyms"].([]interface{}); ok {
		return stringSlice(arr), nil
	}
	if setArr, ok := parsed["synonyms_set"].([]interface{}); ok {
		out := []string{}
		for _, item := range setArr {
			if m, ok := item.(map[string]interface{}); ok {
				if s, ok := m["synonyms"].(string); ok {
					out = append(out, s)
				}
			}
		}
		return out, nil
	}
	return []string{}, nil
}

func (app *App) putSynonymsSet(set string, synonyms []string) error {
	payload := map[string]interface{}{
		"synonyms": synonyms,
	}
	body, _ := json.Marshal(payload)
	respBody, status, err := app.doOpenSearchRequest("PUT", fmt.Sprintf("/_plugins/_synonyms/%s", set), body)
	if err == nil && status < 400 {
		return nil
	}

	payloadAlt := map[string]interface{}{
		"synonyms_set": []map[string]string{},
	}
	for i, s := range synonyms {
		payloadAlt["synonyms_set"] = append(payloadAlt["synonyms_set"].([]map[string]string), map[string]string{
			"id":       fmt.Sprintf("syn-%d", i+1),
			"synonyms": s,
		})
	}
	altBody, _ := json.Marshal(payloadAlt)
	respBody, status, err = app.doOpenSearchRequest("PUT", fmt.Sprintf("/_plugins/_synonyms/%s", set), altBody)
	if err != nil {
		return err
	}
	if status >= 400 {
		return fmt.Errorf("put synonyms failed: %s", string(respBody))
	}
	return nil
}

func (app *App) doOpenSearchRequest(method, path string, body []byte) ([]byte, int, error) {
	if err := app.ensureOpenSearchReady(); err != nil {
		return nil, 0, err
	}
	url := strings.TrimRight(app.OpenSearchURL, "/") + "/" + strings.TrimLeft(path, "/")
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}
	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		return nil, 0, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if app.OpenSearchUsername != "" {
		req.SetBasicAuth(app.OpenSearchUsername, app.OpenSearchPassword)
	}
	resp, err := app.OpenSearchHTTP.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	return respBody, resp.StatusCode, nil
}

func loadOpenSearchMapping() ([]byte, error) {
	path := filepath.Join("scripts", "opensearch_kjv_mapping.json")
	if data, err := os.ReadFile(path); err == nil {
		return data, nil
	}

	fallback := map[string]interface{}{
		"settings": map[string]interface{}{
			"analysis": map[string]interface{}{
				"filter": map[string]interface{}{
					"kjv_synonyms": map[string]interface{}{
						"type":         "synonym_graph",
						"synonyms_set": getEnvDefault("OPENSEARCH_SYNONYMS_SET", defaultSynonymsSet),
					},
				},
				"analyzer": map[string]interface{}{
					"english_with_synonyms": map[string]interface{}{
						"type":      "custom",
						"tokenizer": "standard",
						"filter":    []string{"lowercase", "kjv_synonyms"},
					},
				},
			},
		},
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"book":          map[string]interface{}{"type": "keyword"},
				"chapter":       map[string]interface{}{"type": "integer"},
				"verse":         map[string]interface{}{"type": "integer"},
				"text":          map[string]interface{}{"type": "text", "analyzer": "english", "search_analyzer": "english_with_synonyms"},
				"text_suggest":  map[string]interface{}{"type": "completion"},
				"ordinal_verse": map[string]interface{}{"type": "integer"},
				"ordinal_book":  map[string]interface{}{"type": "integer"},
				"testament":     map[string]interface{}{"type": "keyword"},
			},
		},
	}
	return json.Marshal(fallback)
}

func parseIntDefault(v string, def int) int {
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func jsonError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]interface{}{
		"status":  "error",
		"message": msg,
	})
}

func getEnvDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func stringSlice(items []interface{}) []string {
	out := make([]string, 0, len(items))
	for _, item := range items {
		if s, ok := item.(string); ok {
			out = append(out, s)
		}
	}
	return out
}

func mergeSynonyms(existing, add []string) []string {
	set := map[string]bool{}
	for _, s := range existing {
		set[s] = true
	}
	for _, s := range add {
		if s != "" {
			set[s] = true
		}
	}
	out := make([]string, 0, len(set))
	for s := range set {
		out = append(out, s)
	}
	return out
}

func removeSynonyms(existing, remove []string) []string {
	set := map[string]bool{}
	for _, s := range existing {
		set[s] = true
	}
	for _, s := range remove {
		delete(set, s)
	}
	out := make([]string, 0, len(set))
	for s := range set {
		out = append(out, s)
	}
	return out
}

func muxVar(r *http.Request, key string) string {
	if r == nil {
		return ""
	}
	if v, ok := mux.Vars(r)[key]; ok {
		return v
	}
	return ""
}

// PreloadVerses fetches all verses from OpenSearch using the scroll API
// and populates AllVerses and FlatVerses on the App struct.
// These fields are immutable after this call — safe for concurrent reads.
func (app *App) PreloadVerses() error {
	if err := app.ensureOpenSearchReady(); err != nil {
		return err
	}

	body := map[string]interface{}{
		"query":   map[string]interface{}{"match_all": map[string]interface{}{}},
		"sort":    []interface{}{map[string]interface{}{"ordinal_verse": "asc"}},
		"size":    5000,
		"_source": []string{"book", "chapter", "verse", "text"},
	}
	bodyBytes, _ := json.Marshal(body)

	// Initial scroll request
	respBody, status, err := app.doOpenSearchRequest("POST", fmt.Sprintf("/%s/_search?scroll=1m", app.OpenSearchIndex), bodyBytes)
	if err != nil {
		return fmt.Errorf("scroll init failed: %w", err)
	}
	if status >= 400 {
		return fmt.Errorf("scroll init error: %s", string(respBody))
	}

	type scrollResponse struct {
		ScrollID string `json:"_scroll_id"`
		Hits     struct {
			Hits []struct {
				Source struct {
					Book    string `json:"book"`
					Chapter int    `json:"chapter"`
					Verse   int    `json:"verse"`
					Text    string `json:"text"`
				} `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	var allVerses []Verse
	versesMap := make(map[string]map[int][]Verse)

	addHits := func(resp *scrollResponse) {
		for _, hit := range resp.Hits.Hits {
			v := Verse{
				Book:    hit.Source.Book,
				Chapter: hit.Source.Chapter,
				Verse:   hit.Source.Verse,
				Text:    hit.Source.Text,
			}
			allVerses = append(allVerses, v)
			if versesMap[v.Book] == nil {
				versesMap[v.Book] = make(map[int][]Verse)
			}
			versesMap[v.Book][v.Chapter] = append(versesMap[v.Book][v.Chapter], v)
		}
	}

	// Parse initial response
	var parsed scrollResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return fmt.Errorf("scroll parse failed: %w", err)
	}
	addHits(&parsed)
	scrollID := parsed.ScrollID

	// Follow-up scroll requests
	for len(parsed.Hits.Hits) > 0 {
		scrollBody, _ := json.Marshal(map[string]interface{}{
			"scroll":    "1m",
			"scroll_id": scrollID,
		})
		respBody, status, err = app.doOpenSearchRequest("POST", "/_search/scroll", scrollBody)
		if err != nil || status >= 400 {
			break
		}
		parsed = scrollResponse{}
		if err := json.Unmarshal(respBody, &parsed); err != nil {
			break
		}
		if len(parsed.Hits.Hits) == 0 {
			break
		}
		addHits(&parsed)
		scrollID = parsed.ScrollID
	}

	// Clear scroll
	clearBody, _ := json.Marshal(map[string]interface{}{"scroll_id": scrollID})
	app.doOpenSearchRequest("DELETE", "/_search/scroll", clearBody)

	if len(allVerses) == 0 {
		return fmt.Errorf("no verses found in index %s", app.OpenSearchIndex)
	}

	app.AllVerses = versesMap
	app.FlatVerses = allVerses
	return nil
}

// OpenSearch data access helpers for v1 handler migration

type osHitsResponse struct {
	Hits struct {
		Total struct {
			Value int `json:"value"`
		} `json:"total"`
		Hits []struct {
			Source struct {
				Book        string `json:"book"`
				Chapter     int    `json:"chapter"`
				Verse       int    `json:"verse"`
				Text        string `json:"text"`
				OrdinalBook int    `json:"ordinal_book"`
			} `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
	Aggregations struct {
		ByBook struct {
			Buckets []struct {
				Key      string `json:"key"`
				DocCount int    `json:"doc_count"`
			} `json:"buckets"`
		} `json:"by_book"`
	} `json:"aggregations"`
}

func (app *App) osGetChapter(book string, chapter int) ([]Verse, error) {
	if err := app.ensureOpenSearchReady(); err != nil {
		return nil, err
	}

	body := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"filter": []interface{}{
					map[string]interface{}{"term": map[string]interface{}{"book": book}},
					map[string]interface{}{"term": map[string]interface{}{"chapter": chapter}},
				},
			},
		},
		"sort":    []interface{}{map[string]interface{}{"verse": "asc"}},
		"size":    200,
		"_source": []string{"verse", "text"},
	}
	bodyBytes, _ := json.Marshal(body)

	respBody, status, err := app.doOpenSearchRequest("POST", fmt.Sprintf("/%s/_search", app.OpenSearchIndex), bodyBytes)
	if err != nil {
		return nil, fmt.Errorf("opensearch request failed: %w", err)
	}
	if status >= 400 {
		return nil, fmt.Errorf("opensearch error: %s", string(respBody))
	}

	var parsed osHitsResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return nil, fmt.Errorf("failed to parse opensearch response: %w", err)
	}

	verses := make([]Verse, 0, len(parsed.Hits.Hits))
	for _, hit := range parsed.Hits.Hits {
		verses = append(verses, Verse{
			Book:    hit.Source.Book,
			Chapter: hit.Source.Chapter,
			Verse:   hit.Source.Verse,
			Text:    hit.Source.Text,
		})
	}
	return verses, nil
}

func (app *App) osGetVerses(book string, chapter int, verseNums []int) ([]Verse, error) {
	if err := app.ensureOpenSearchReady(); err != nil {
		return nil, err
	}

	body := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"filter": []interface{}{
					map[string]interface{}{"term": map[string]interface{}{"book": book}},
					map[string]interface{}{"term": map[string]interface{}{"chapter": chapter}},
					map[string]interface{}{"terms": map[string]interface{}{"verse": verseNums}},
				},
			},
		},
		"sort":    []interface{}{map[string]interface{}{"verse": "asc"}},
		"size":    200,
		"_source": []string{"verse", "text"},
	}
	bodyBytes, _ := json.Marshal(body)

	respBody, status, err := app.doOpenSearchRequest("POST", fmt.Sprintf("/%s/_search", app.OpenSearchIndex), bodyBytes)
	if err != nil {
		return nil, fmt.Errorf("opensearch request failed: %w", err)
	}
	if status >= 400 {
		return nil, fmt.Errorf("opensearch error: %s", string(respBody))
	}

	var parsed osHitsResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return nil, fmt.Errorf("failed to parse opensearch response: %w", err)
	}

	verses := make([]Verse, 0, len(parsed.Hits.Hits))
	for _, hit := range parsed.Hits.Hits {
		verses = append(verses, Verse{
			Book:    hit.Source.Book,
			Chapter: hit.Source.Chapter,
			Verse:   hit.Source.Verse,
			Text:    hit.Source.Text,
		})
	}
	return verses, nil
}

func (app *App) osGetRandomVerse() (Verse, error) {
	if err := app.ensureOpenSearchReady(); err != nil {
		return Verse{}, err
	}

	body := map[string]interface{}{
		"query": map[string]interface{}{
			"function_score": map[string]interface{}{
				"query":        map[string]interface{}{"match_all": map[string]interface{}{}},
				"random_score": map[string]interface{}{},
			},
		},
		"size":    1,
		"_source": []string{"book", "chapter", "verse", "text"},
	}
	bodyBytes, _ := json.Marshal(body)

	respBody, status, err := app.doOpenSearchRequest("POST", fmt.Sprintf("/%s/_search", app.OpenSearchIndex), bodyBytes)
	if err != nil {
		return Verse{}, fmt.Errorf("opensearch request failed: %w", err)
	}
	if status >= 400 {
		return Verse{}, fmt.Errorf("opensearch error: %s", string(respBody))
	}

	var parsed osHitsResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return Verse{}, fmt.Errorf("failed to parse opensearch response: %w", err)
	}

	if len(parsed.Hits.Hits) == 0 {
		return Verse{}, fmt.Errorf("no verses found")
	}

	hit := parsed.Hits.Hits[0]
	return Verse{
		Book:    hit.Source.Book,
		Chapter: hit.Source.Chapter,
		Verse:   hit.Source.Verse,
		Text:    hit.Source.Text,
	}, nil
}

func (app *App) osSearch(query string, limit int) ([]Verse, [66]int, error) {
	if err := app.ensureOpenSearchReady(); err != nil {
		return nil, [66]int{}, err
	}

	body := map[string]interface{}{
		"query": map[string]interface{}{
			"match": map[string]interface{}{
				"text": map[string]interface{}{
					"query": query,
				},
			},
		},
		"size":             limit,
		"track_total_hits": true,
		"_source":          []string{"book", "chapter", "verse", "text", "ordinal_book"},
		"aggs": map[string]interface{}{
			"by_book": map[string]interface{}{
				"terms": map[string]interface{}{
					"field": "book",
					"size":  66,
				},
			},
		},
	}
	bodyBytes, _ := json.Marshal(body)

	respBody, status, err := app.doOpenSearchRequest("POST", fmt.Sprintf("/%s/_search", app.OpenSearchIndex), bodyBytes)
	if err != nil {
		return nil, [66]int{}, fmt.Errorf("opensearch request failed: %w", err)
	}
	if status >= 400 {
		return nil, [66]int{}, fmt.Errorf("opensearch error: %s", string(respBody))
	}

	var parsed osHitsResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return nil, [66]int{}, fmt.Errorf("failed to parse opensearch response: %w", err)
	}

	verses := make([]Verse, 0, len(parsed.Hits.Hits))
	for _, hit := range parsed.Hits.Hits {
		verses = append(verses, Verse{
			Book:    hit.Source.Book,
			Chapter: hit.Source.Chapter,
			Verse:   hit.Source.Verse,
			Text:    hit.Source.Text,
		})
	}

	// Build per-book graph counter from aggregation
	var graphCounter [66]int
	bookToIndex := make(map[string]int, 66)
	for i, name := range BooksCanonicalOrder {
		bookToIndex[name] = i
	}
	for _, bucket := range parsed.Aggregations.ByBook.Buckets {
		if idx, ok := bookToIndex[bucket.Key]; ok {
			graphCounter[idx] = bucket.DocCount
		}
	}

	return verses, graphCounter, nil
}

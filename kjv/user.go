package kjv

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
)

const maxHistoryEntries = 100

// userSettings holds the per-user preferences that sync across devices.
// Mirrors the client-side settings currently kept only in localStorage.
type userSettings struct {
	Theme         string `json:"theme,omitempty"`
	Font          string `json:"font,omitempty"`
	VerseOpenMode string `json:"verse_open_mode,omitempty"`
}

// readPage is one entry in a user's reading history.
type readPage struct {
	Book    string `json:"book"`
	Chapter int    `json:"chapter"`
	TS      int64  `json:"ts"`
}

// searchHistory is one completed full-text search for a user.
type searchHistory struct {
	Query string `json:"query"`
	TS    int64  `json:"ts"`
}

func (app *App) SetupUserRoutes() {
	// Authenticated via the session cookie (see getSession), same as /auth/me.
	// No JWT scope is required: any logged-in user manages their own data.
	app.Router.HandleFunc("/user/settings", app.getUserSettings).Methods("GET")
	app.Router.HandleFunc("/user/settings", app.putUserSettings).Methods("PUT")
	app.Router.HandleFunc("/user/history", app.getUserHistory).Methods("GET")
	app.Router.HandleFunc("/user/history", app.postUserHistory).Methods("POST")
	app.Router.HandleFunc("/user/history/searches", app.postUserSearchHistory).Methods("POST")
}

// currentSub returns the authenticated user's Google subject, or false if the
// request has no valid session.
func (app *App) currentSub(r *http.Request) (string, bool) {
	session, err := app.getSession(r)
	if err != nil || session.Sub == "" {
		return "", false
	}
	return session.Sub, true
}

func settingsKey(sub string) string      { return "user:" + sub + ":settings" }
func historyKey(sub string) string       { return "user:" + sub + ":history" }
func searchHistoryKey(sub string) string { return "user:" + sub + ":search-history" }

func (app *App) getUserSettings(w http.ResponseWriter, r *http.Request) {
	sub, ok := app.currentSub(r)
	if !ok {
		jsonError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	val, err := app.Redis.Get(context.Background(), settingsKey(sub)).Result()
	if err == redis.Nil {
		// No settings saved yet; return an empty object so the client keeps its defaults.
		writeJSON(w, http.StatusOK, userSettings{})
		return
	}
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "failed to load settings")
		return
	}
	var s userSettings
	if err := json.Unmarshal([]byte(val), &s); err != nil {
		jsonError(w, http.StatusInternalServerError, "corrupt settings")
		return
	}
	writeJSON(w, http.StatusOK, s)
}

func (app *App) putUserSettings(w http.ResponseWriter, r *http.Request) {
	sub, ok := app.currentSub(r)
	if !ok {
		jsonError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var s userSettings
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	payload, _ := json.Marshal(s)
	// No TTL: settings persist until the user changes them.
	if err := app.Redis.Set(context.Background(), settingsKey(sub), payload, 0).Err(); err != nil {
		jsonError(w, http.StatusInternalServerError, "failed to save settings")
		return
	}
	writeJSON(w, http.StatusOK, s)
}

func (app *App) getUserHistory(w http.ResponseWriter, r *http.Request) {
	sub, ok := app.currentSub(r)
	if !ok {
		jsonError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	pages, err := app.loadHistory(sub)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "failed to load history")
		return
	}
	searches, err := app.loadSearchHistory(sub)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "failed to load search history")
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"pages": pages, "searches": searches})
}

func (app *App) postUserHistory(w http.ResponseWriter, r *http.Request) {
	sub, ok := app.currentSub(r)
	if !ok {
		jsonError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req struct {
		Book    string `json:"book"`
		Chapter int    `json:"chapter"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	book := strings.ToUpper(strings.TrimSpace(req.Book))
	if book == "" || req.Chapter <= 0 {
		jsonError(w, http.StatusBadRequest, "book and chapter are required")
		return
	}

	pages, err := app.loadHistory(sub)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "failed to load history")
		return
	}

	// Drop any prior entry for this exact page so it moves to the front
	// instead of appearing twice.
	filtered := pages[:0]
	for _, p := range pages {
		if p.Book == book && p.Chapter == req.Chapter {
			continue
		}
		filtered = append(filtered, p)
	}
	// Prepend the just-read page (most-recent-first) and cap the list.
	updated := append([]readPage{{Book: book, Chapter: req.Chapter, TS: time.Now().Unix()}}, filtered...)
	if len(updated) > maxHistoryEntries {
		updated = updated[:maxHistoryEntries]
	}

	payload, _ := json.Marshal(updated)
	if err := app.Redis.Set(context.Background(), historyKey(sub), payload, 0).Err(); err != nil {
		jsonError(w, http.StatusInternalServerError, "failed to save history")
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"pages": updated})
	if app.Analytics != nil {
		app.Analytics.RecordActivity(r, sub, "chapter_read", map[string]interface{}{"book": book, "chapter": req.Chapter})
	}
}

func (app *App) postUserSearchHistory(w http.ResponseWriter, r *http.Request) {
	sub, ok := app.currentSub(r)
	if !ok {
		jsonError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req struct {
		Query string `json:"query"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	query := strings.TrimSpace(req.Query)
	if query == "" {
		jsonError(w, http.StatusBadRequest, "query is required")
		return
	}

	searches, err := app.loadSearchHistory(sub)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "failed to load search history")
		return
	}
	filtered := searches[:0]
	for _, entry := range searches {
		if strings.EqualFold(entry.Query, query) {
			continue
		}
		filtered = append(filtered, entry)
	}
	updated := append([]searchHistory{{Query: query, TS: time.Now().Unix()}}, filtered...)
	if len(updated) > maxHistoryEntries {
		updated = updated[:maxHistoryEntries]
	}

	payload, _ := json.Marshal(updated)
	if err := app.Redis.Set(context.Background(), searchHistoryKey(sub), payload, 0).Err(); err != nil {
		jsonError(w, http.StatusInternalServerError, "failed to save search history")
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"searches": updated})
	if app.Analytics != nil {
		app.Analytics.RecordActivity(r, sub, "search", map[string]string{"query": query})
	}
}

// loadHistory returns the user's reading history, most-recent-first, or an
// empty slice if none is stored yet.
func (app *App) loadHistory(sub string) ([]readPage, error) {
	val, err := app.Redis.Get(context.Background(), historyKey(sub)).Result()
	if err == redis.Nil {
		return []readPage{}, nil
	}
	if err != nil {
		return nil, err
	}
	var pages []readPage
	if err := json.Unmarshal([]byte(val), &pages); err != nil {
		// Treat corrupt history as empty rather than failing the request.
		return []readPage{}, nil
	}
	return pages, nil
}

func (app *App) loadSearchHistory(sub string) ([]searchHistory, error) {
	val, err := app.Redis.Get(context.Background(), searchHistoryKey(sub)).Result()
	if err == redis.Nil {
		return []searchHistory{}, nil
	}
	if err != nil {
		return nil, err
	}
	var searches []searchHistory
	if err := json.Unmarshal([]byte(val), &searches); err != nil {
		return []searchHistory{}, nil
	}
	return searches, nil
}

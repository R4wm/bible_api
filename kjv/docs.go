package kjv

import (
	"net/http"
)

type EndpointDoc struct {
	Method      string `json:"method"`
	Path        string `json:"path"`
	Description string `json:"description"`
	Auth        string `json:"auth,omitempty"`
}

func (app *App) SetupDocsRoutes() {
	app.Router.HandleFunc("/docs", app.docsPage).Methods("GET")
	app.Router.HandleFunc("/docs.json", app.docsJSON).Methods("GET")
}

func (app *App) docsJSON(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":    "ok",
		"endpoints": allEndpoints(),
	})
}

func (app *App) docsPage(w http.ResponseWriter, r *http.Request) {
	t, err := app.docsTemplate()
	if err != nil {
		http.Error(w, "Failed to render docs", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := t.Execute(w, allEndpoints()); err != nil {
		http.Error(w, "Failed to render docs", http.StatusInternalServerError)
	}
}

func allEndpoints() []EndpointDoc {
	return []EndpointDoc{
		{Method: "GET", Path: "/bible/list_books", Description: "List all Bible books"},
		{Method: "GET", Path: "/bible/list_chapters/{book}", Description: "List chapters in a book"},
		{Method: "GET", Path: "/bible/{book}", Description: "List chapters in a book"},
		{Method: "GET", Path: "/bible/{book}/{chapter}", Description: "Get all verses in a chapter"},
		{Method: "GET", Path: "/bible/{book}/{chapter}/{verse}", Description: "Get specific verse"},
		{Method: "GET", Path: "/bible/{book}/{chapter}/{start-end}", Description: "Get verse range"},
		{Method: "GET", Path: "/bible/search?q={query}", Description: "Search Bible text"},
		{Method: "GET", Path: "/bible/random_verse", Description: "Get random verse"},
		{Method: "GET", Path: "/bible/v2/search?q={query}&match={any|all|phrase}&case_sensitive={true|false}", Description: "OpenSearch full-text search. match defaults to any; case_sensitive defaults to false and requires the case-preserving index mapping."},
		{Method: "GET", Path: "/bible/v2/suggest?q={prefix}", Description: "Predictive suggestions"},
		{Method: "PUT", Path: "/bible/v2/synonyms/{set}", Description: "Replace synonym set", Auth: "JWT (scope: synonyms:write)"},
		{Method: "POST", Path: "/bible/v2/synonyms/{set}", Description: "Append to synonym set", Auth: "JWT (scope: synonyms:write)"},
		{Method: "DELETE", Path: "/bible/v2/synonyms/{set}", Description: "Remove from synonym set", Auth: "JWT (scope: synonyms:write)"},
		{Method: "POST", Path: "/auth/google/token", Description: "Exchange Google ID token for app JWT + session"},
		{Method: "GET", Path: "/auth/config", Description: "Auth configuration (Google client id)"},
		{Method: "GET", Path: "/auth/me", Description: "Get current session token"},
		{Method: "POST", Path: "/auth/logout", Description: "Clear session"},
		{Method: "POST", Path: "/admin/token", Description: "Mint token with X-Internal-Secret", Auth: "X-Internal-Secret header"},
		{Method: "GET", Path: "/admin/rate-limit/{ip}", Description: "Check IP rate limit status"},
		{Method: "POST", Path: "/admin/block-ip", Description: "Manually block an IP"},
		{Method: "DELETE", Path: "/admin/unblock-ip/{ip}", Description: "Unblock an IP"},
		{Method: "GET", Path: "/admin/blocked-ips", Description: "List blocked IPs"},
		{Method: "GET", Path: "/health", Description: "Health check"},
		{Method: "GET", Path: "/v2", Description: "Web UI (Google login + search/suggest)"},
		{Method: "GET", Path: "/docs", Description: "API documentation (HTML)"},
		{Method: "GET", Path: "/docs.json", Description: "API documentation (JSON)"},
	}
}

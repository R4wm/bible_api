package kjv

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3"
)

// setupTestApp creates a test app instance with an in-memory database
func setupTestApp(t *testing.T) *App {
	// Create in-memory SQLite database for testing
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Create a simple test table with a few verses for search testing
	_, err = db.Exec(`
		CREATE TABLE kjv (
			book TEXT,
			chapter INTEGER,
			verse INTEGER,
			text TEXT,
			ordinal_verse INTEGER,
			ordinal_book INTEGER
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create test table: %v", err)
	}

	// Insert test data
	_, err = db.Exec(`
		INSERT INTO kjv (book, chapter, verse, text, ordinal_verse, ordinal_book) VALUES
		('JOHN', 3, 16, 'For God so loved the world, that he gave his only begotten Son', 1, 1),
		('1JOHN', 4, 8, 'He that loveth not knoweth not God; for God is love', 2, 2),
		('ROMANS', 8, 28, 'And we know that all things work together for good to them that love God', 3, 3)
	`)
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	router := mux.NewRouter()
	app := &App{
		Router:   router,
		Database: db,
		Redis:    nil, // No Redis needed for these tests
	}

	// Setup only the search route for testing
	app.Router.HandleFunc("/bible/search", app.search)

	return app
}

func TestSearchWithoutLimit(t *testing.T) {
	app := setupTestApp(t)
	defer app.Database.Close()

	// Create request without limit parameter
	req, err := http.NewRequest("GET", "/bible/search?q=love", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call the handler
	app.Router.ServeHTTP(rr, req)

	// Check status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status code %d, got %d. Response: %s", http.StatusOK, status, rr.Body.String())
	}

	// Check that we get results (should contain "love" in the response)
	responseBody := rr.Body.String()
	if !strings.Contains(responseBody, "love") {
		t.Errorf("Expected response to contain 'love', but got: %s", responseBody)
	}
	
	// Check that JSON book breakdown is present
	if !strings.Contains(responseBody, "Book breakdown (JSON):") {
		t.Errorf("Expected response to contain JSON book breakdown section, but got: %s", responseBody)
	}
	
	// Check that JSON breakdown contains book counts
	if !strings.Contains(responseBody, "JOHN\": 1") {
		t.Errorf("Expected response to contain JSON book counts, but got: %s", responseBody)
	}
}

func TestSearchWithLimit(t *testing.T) {
	app := setupTestApp(t)
	defer app.Database.Close()

	// Create request with limit parameter
	req, err := http.NewRequest("GET", "/bible/search?q=love&n=2", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call the handler
	app.Router.ServeHTTP(rr, req)

	// Check status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status code %d, got %d. Response: %s", http.StatusOK, status, rr.Body.String())
	}

	// Check that we get results
	responseBody := rr.Body.String()
	if !strings.Contains(responseBody, "love") {
		t.Errorf("Expected response to contain 'love', but got: %s", responseBody)
	}
}

func TestSearchInvalidLimit(t *testing.T) {
	app := setupTestApp(t)
	defer app.Database.Close()

	// Test invalid limit (too high)
	req, err := http.NewRequest("GET", "/bible/search?q=love&n=20000", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	app.Router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("Expected status code %d for invalid limit, got %d", http.StatusBadRequest, status)
	}

	expectedError := "Search limit must be a number between 1 and 10000"
	if !strings.Contains(rr.Body.String(), expectedError) {
		t.Errorf("Expected error message about limit, got: %s", rr.Body.String())
	}
}

func TestSearchInvalidCharacters(t *testing.T) {
	app := setupTestApp(t)
	defer app.Database.Close()

	// Test search with invalid characters
	req, err := http.NewRequest("GET", "/bible/search?q=love<script>", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	app.Router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("Expected status code %d for invalid characters, got %d", http.StatusBadRequest, status)
	}

	expectedError := "Search string contains invalid characters"
	if !strings.Contains(rr.Body.String(), expectedError) {
		t.Errorf("Expected error message about invalid characters, got: %s", rr.Body.String())
	}
}

func TestSearchTooShort(t *testing.T) {
	app := setupTestApp(t)
	defer app.Database.Close()

	// Test search string too short
	req, err := http.NewRequest("GET", "/bible/search?q=a", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	app.Router.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("Expected status code %d for short search string, got %d", http.StatusBadRequest, status)
	}

	expectedError := "Search string must be between 2 and 100 characters"
	if !strings.Contains(rr.Body.String(), expectedError) {
		t.Errorf("Expected error message about string length, got: %s", rr.Body.String())
	}
}

func TestSearchMissingQuery(t *testing.T) {
	app := setupTestApp(t)
	defer app.Database.Close()

	// Test missing query parameter
	req, err := http.NewRequest("GET", "/bible/search", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	app.Router.ServeHTTP(rr, req)

	// Should return the search form
	responseBody := rr.Body.String()
	if !strings.Contains(responseBody, "Search the Bible") {
		t.Errorf("Expected search form for missing query, got: %s", responseBody)
	}
	if !strings.Contains(responseBody, "Popular searches:") {
		t.Errorf("Expected popular searches in search form, got: %s", responseBody)
	}
}

func TestWantsJson(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected bool
	}{
		{"JSON true", "/test?json=true", true},
		{"JSON false", "/test?json=false", false},
		{"No JSON param", "/test", false},
		{"JSON empty", "/test?json=", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", tt.url, nil)
			result := wantsJson(req)
			if result != tt.expected {
				t.Errorf("wantsJson() = %v, expected %v for URL %s", result, tt.expected, tt.url)
			}
		})
	}
}

func TestLazyBook(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    string
		shouldError bool
	}{
		{"Exact match", "JOHN", "JOHN", false},
		{"Partial match", "JOH", "JOHN", false},
		{"Case insensitive", "john", "JOHN", false},
		{"Multiple matches", "1", "", true}, // Should match 1SAMUEL, 1KINGS, etc.
		{"No match", "INVALIDBOOK", "", true},
		{"Empty string", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := lazyBook(tt.input)
			
			if tt.shouldError {
				if err == nil {
					t.Errorf("lazyBook(%s) expected error but got none", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("lazyBook(%s) unexpected error: %v", tt.input, err)
				}
				if result != tt.expected {
					t.Errorf("lazyBook(%s) = %s, expected %s", tt.input, result, tt.expected)
				}
			}
		})
	}
}
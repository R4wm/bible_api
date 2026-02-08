package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRemoveTrailingSlash(t *testing.T) {
	called := false
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		if r.URL.Path != "/foo" {
			t.Fatalf("expected path /foo, got %s", r.URL.Path)
		}
	})

	req := httptest.NewRequest(http.MethodGet, "http://example.com/foo/", nil)
	rr := httptest.NewRecorder()
	removeTrailingSlash(h).ServeHTTP(rr, req)

	if !called {
		t.Fatal("handler was not called")
	}
}

func TestRemoveTrailingSlashNoChange(t *testing.T) {
	called := false
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		if r.URL.Path != "/bar" {
			t.Fatalf("expected path /bar, got %s", r.URL.Path)
		}
	})

	req := httptest.NewRequest(http.MethodGet, "http://example.com/bar", nil)
	rr := httptest.NewRecorder()
	removeTrailingSlash(h).ServeHTTP(rr, req)

	if !called {
		t.Fatal("handler was not called")
	}
}

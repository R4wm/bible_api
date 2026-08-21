package kjv

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
)

func TestMetricsEndpointAndMiddleware(t *testing.T) {
	analytics := NewAnalytics("")
	router := mux.NewRouter()
	analytics.WrapRouter(router)
	router.HandleFunc("/bible/{book}", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}).Methods(http.MethodGet)
	router.HandleFunc("/fail", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}).Methods(http.MethodGet)
	router.Handle("/metrics", analytics.MetricsHandler()).Methods(http.MethodGet)

	router.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/bible/John", nil))
	router.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/fail", nil))
	router.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/missing", nil))
	analytics.RecordChapterServed("John", 3)
	analytics.RecordChapterServed("PSALMS", 23)
	analytics.RecordChapterServed("NotABook", 1)
	analytics.RecordSearch("v2", "ok", 12*time.Millisecond)
	analytics.RecordSearch("v2", "invalid", time.Millisecond)

	metricsRequest := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	metricsResponse := httptest.NewRecorder()
	router.ServeHTTP(metricsResponse, metricsRequest)

	if metricsResponse.Code != http.StatusOK {
		t.Fatalf("metrics status = %d, want 200", metricsResponse.Code)
	}
	body := metricsResponse.Body.String()
	for _, metric := range []string{
		"go_goroutines",
		"bible_api_http_requests_total{method=\"GET\",route=\"/bible/{book}\",status=\"200\",status_class=\"2xx\"}",
		"bible_api_http_requests_total{method=\"GET\",route=\"/fail\",status=\"500\",status_class=\"5xx\"}",
		"bible_api_http_requests_total{method=\"GET\",route=\"unmatched\",status=\"404\",status_class=\"4xx\"}",
		"bible_api_http_request_duration_seconds",
		"bible_api_http_in_flight_requests",
		"bible_api_chapters_served_total{book=\"JOHN\",chapter=\"3\"}",
		"bible_api_chapters_served_total{book=\"PSALMS\",chapter=\"23\"}",
		"bible_api_search_requests_total{api=\"v2\",result=\"ok\"}",
		"bible_api_search_requests_total{api=\"v2\",result=\"invalid\"}",
		"bible_api_search_duration_seconds",
	} {
		if !strings.Contains(body, metric) {
			t.Errorf("metrics output missing %q", metric)
		}
	}
	if strings.Contains(body, `bible_api_chapters_served_total{book="NOTABOOK"`) {
		t.Errorf("unknown book was recorded as a chapter metric")
	}
}

func TestCanonicalServedChapter(t *testing.T) {
	book, chapter, ok := canonicalServedChapter("john", 3)
	if !ok || book != "JOHN" || chapter != "3" {
		t.Fatalf("got book=%q chapter=%q ok=%v", book, chapter, ok)
	}
	if _, _, ok := canonicalServedChapter("JOHN", 0); ok {
		t.Fatal("chapter 0 should be rejected")
	}
	if _, _, ok := canonicalServedChapter("JOHN", 22); ok {
		t.Fatal("John has 21 chapters")
	}
	if _, _, ok := canonicalServedChapter("Nope", 1); ok {
		t.Fatal("unknown book should be rejected")
	}
}

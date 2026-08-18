package kjv

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
)

func TestMetricsEndpointAndMiddleware(t *testing.T) {
	analytics := NewAnalytics("")
	router := mux.NewRouter()
	router.Use(analytics.HTTPMiddleware)
	router.HandleFunc("/bible/{book}", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}).Methods(http.MethodGet)
	router.Handle("/metrics", analytics.MetricsHandler()).Methods(http.MethodGet)

	request := httptest.NewRequest(http.MethodGet, "/bible/John", nil)
	router.ServeHTTP(httptest.NewRecorder(), request)

	metricsRequest := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	metricsResponse := httptest.NewRecorder()
	router.ServeHTTP(metricsResponse, metricsRequest)

	if metricsResponse.Code != http.StatusOK {
		t.Fatalf("metrics status = %d, want 200", metricsResponse.Code)
	}
	body := metricsResponse.Body.String()
	for _, metric := range []string{
		"go_goroutines",
		"bible_api_http_requests_total{method=\"GET\",route=\"/bible/{book}\",status_class=\"2xx\"}",
		"bible_api_http_request_duration_seconds",
		"bible_api_http_in_flight_requests",
	} {
		if !strings.Contains(body, metric) {
			t.Errorf("metrics output missing %q", metric)
		}
	}
}

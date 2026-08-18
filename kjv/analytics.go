package kjv

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const analyticsRetention = 3 * 365 * 24 * time.Hour

type Analytics struct {
	db       *sql.DB
	requests *prometheus.CounterVec
	duration *prometheus.HistogramVec
	inFlight prometheus.Gauge
	events   *prometheus.CounterVec
}

func NewAnalytics(databaseURL string) *Analytics {
	a := &Analytics{
		requests: prometheus.NewCounterVec(prometheus.CounterOpts{Name: "bible_api_http_requests_total", Help: "HTTP requests served."}, []string{"method", "route", "status_class"}),
		duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{Name: "bible_api_http_request_duration_seconds", Help: "HTTP request duration."}, []string{"method", "route"}),
		inFlight: prometheus.NewGauge(prometheus.GaugeOpts{Name: "bible_api_http_in_flight_requests", Help: "HTTP requests currently in flight."}),
		events:   prometheus.NewCounterVec(prometheus.CounterOpts{Name: "bible_api_activity_events_total", Help: "Durable user activity events."}, []string{"type"}),
	}
	prometheus.MustRegister(a.requests, a.duration, a.inFlight, a.events)
	if strings.TrimSpace(databaseURL) == "" {
		return a
	}
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		log.Printf("analytics database disabled: %v", err)
		return a
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		log.Printf("analytics database unavailable: %v", err)
		return a
	}
	a.db = db
	if err := a.migrate(ctx); err != nil {
		log.Printf("analytics migration failed: %v", err)
		a.db = nil
	}
	return a
}

func (a *Analytics) migrate(ctx context.Context) error {
	_, err := a.db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS users (sub TEXT PRIMARY KEY, email TEXT, name TEXT, first_seen TIMESTAMPTZ NOT NULL DEFAULT now(), last_seen TIMESTAMPTZ NOT NULL DEFAULT now());
CREATE TABLE IF NOT EXISTS auth_events (id BIGSERIAL PRIMARY KEY, sub TEXT, email TEXT, name TEXT, event_type TEXT NOT NULL, ip TEXT, user_agent TEXT, occurred_at TIMESTAMPTZ NOT NULL DEFAULT now());
CREATE TABLE IF NOT EXISTS activity_events (id BIGSERIAL PRIMARY KEY, sub TEXT, event_type TEXT NOT NULL, details JSONB NOT NULL DEFAULT '{}'::jsonb, ip TEXT, user_agent TEXT, occurred_at TIMESTAMPTZ NOT NULL DEFAULT now());
CREATE INDEX IF NOT EXISTS activity_events_sub_time_idx ON activity_events (sub, occurred_at DESC);
CREATE INDEX IF NOT EXISTS auth_events_sub_time_idx ON auth_events (sub, occurred_at DESC);`)
	return err
}

func clientIP(r *http.Request) string {
	if x := strings.TrimSpace(strings.Split(r.Header.Get("X-Forwarded-For"), ",")[0]); x != "" {
		return x
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}

func (a *Analytics) RecordAuth(r *http.Request, s sessionData, event string) {
	a.events.WithLabelValues(event).Inc()
	if a.db == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, err := a.db.ExecContext(ctx, `INSERT INTO users (sub,email,name) VALUES ($1,$2,$3) ON CONFLICT (sub) DO UPDATE SET email=EXCLUDED.email,name=EXCLUDED.name,last_seen=now()`, s.Sub, s.Email, s.Name)
	if err == nil {
		_, err = a.db.ExecContext(ctx, `INSERT INTO auth_events (sub,email,name,event_type,ip,user_agent) VALUES ($1,$2,$3,$4,$5,$6)`, s.Sub, s.Email, s.Name, event, clientIP(r), r.UserAgent())
	}
	if err != nil {
		log.Printf("analytics auth write failed: %v", err)
	}
}

func (a *Analytics) RecordActivity(r *http.Request, sub, event string, details any) {
	a.events.WithLabelValues(event).Inc()
	if a.db == nil {
		return
	}
	b, _ := json.Marshal(details)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_, err := a.db.ExecContext(ctx, `INSERT INTO activity_events (sub,event_type,details,ip,user_agent) VALUES ($1,$2,$3::jsonb,$4,$5)`, sub, event, string(b), clientIP(r), r.UserAgent())
	if err != nil {
		log.Printf("analytics activity write failed: %v", err)
	}
}

func (a *Analytics) Cleanup() {
	if a.db == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cutoff := time.Now().Add(-analyticsRetention)
	_, _ = a.db.ExecContext(ctx, `DELETE FROM auth_events WHERE occurred_at < $1; DELETE FROM activity_events WHERE occurred_at < $1; DELETE FROM users WHERE last_seen < $1`, cutoff)
}

type metricsResponseWriter struct {
	http.ResponseWriter
	status int
}

func (w *metricsResponseWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *metricsResponseWriter) Write(body []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	return w.ResponseWriter.Write(body)
}

// HTTPMiddleware records cardinality-safe request metrics. It uses Gorilla's
// matched route template rather than raw request paths, excluding /metrics so
// Prometheus scrapes cannot inflate application traffic dashboards.
func (a *Analytics) HTTPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/metrics" {
			next.ServeHTTP(w, r)
			return
		}

		started := time.Now()
		a.inFlight.Inc()
		defer a.inFlight.Dec()

		recorder := &metricsResponseWriter{ResponseWriter: w}
		next.ServeHTTP(recorder, r)
		if recorder.status == 0 {
			recorder.status = http.StatusOK
		}

		route := "unmatched"
		if current := mux.CurrentRoute(r); current != nil {
			if template, err := current.GetPathTemplate(); err == nil && template != "" {
				route = template
			}
		}
		statusClass := fmt.Sprintf("%dxx", recorder.status/100)
		a.requests.WithLabelValues(r.Method, route, statusClass).Inc()
		a.duration.WithLabelValues(r.Method, route).Observe(time.Since(started).Seconds())
	})
}

func (a *Analytics) MetricsHandler() http.Handler { return promhttp.Handler() }

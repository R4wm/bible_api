package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"github.com/r4wm/bible_api/kjv"
	"github.com/r4wm/bible_api/middleware"
	log "github.com/sirupsen/logrus"
)

func removeTrailingSlash(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = strings.TrimSuffix(r.URL.Path, "/")
		next.ServeHTTP(w, r)
	})
}

func waitForOpenSearch(url string, client *http.Client, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	healthURL := strings.TrimRight(url, "/") + "/_cluster/health?wait_for_status=yellow&timeout=5s"

	for time.Now().Before(deadline) {
		resp, err := client.Get(healthURL)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}
		log.Infof("Waiting for OpenSearch at %s...", url)
		time.Sleep(2 * time.Second)
	}
	return fmt.Errorf("opensearch not ready at %s after %v", url, timeout)
}

func main() {

	// Initialize Redis client
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	redisPassword := os.Getenv("REDIS_PASSWORD")

	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       0, // Default DB
	})

	// Test Redis connection
	ctx := context.Background()
	pong, err := rdb.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Could not connect to Redis: %v", err)
	}
	log.Infof("Connected to Redis: %s", pong)

	// Router
	router := mux.NewRouter().StrictSlash(false)

	corsOrigins := strings.Split(getEnvOrDefault("CORS_ALLOW_ORIGINS", "http://localhost:5173,http://127.0.0.1:5173"), ",")
	router.Use(middleware.CORSMiddleware(corsOrigins))

	// Create rate limiter middleware
	rateLimiter := middleware.NewRateLimiter(rdb)

	// Apply rate limiting middleware to all routes
	router.Use(rateLimiter.Middleware)

	// Add health check endpoint that bypasses rate limiting
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "healthy", "service": "bible_api", "redis": "connected"}`))
	}).Methods("GET")

	app := kjv.App{
		Router: router,
		Redis:  rdb,
	}
	app.OpenSearchURL = getEnvOrDefault("OPENSEARCH_URL", "http://localhost:9200")
	app.OpenSearchIndex = getEnvOrDefault("OPENSEARCH_INDEX", "kjv_v2")
	app.OpenSearchUsername = os.Getenv("OPENSEARCH_USERNAME")
	app.OpenSearchPassword = os.Getenv("OPENSEARCH_PASSWORD")
	app.OpenSearchHTTP = &http.Client{Timeout: 10 * time.Second}

	app.JWTSecret = []byte(os.Getenv("JWT_SECRET"))
	app.JWTIssuer = getEnvOrDefault("JWT_ISSUER", "bible_api")
	app.JWTAudience = getEnvOrDefault("JWT_AUDIENCE", "bible_api_clients")
	app.JWTTTLSeconds = getEnvInt("JWT_TTL_SECONDS", 3600)
	app.SessionTTLSeconds = getEnvInt("SESSION_TTL_SECONDS", 3600)
	app.SessionCookieName = getEnvOrDefault("SESSION_COOKIE_NAME", "bible_api_session")
	app.SessionCookieSecure = parseBoolEnv(os.Getenv("SESSION_COOKIE_SECURE"), false)
	app.GoogleClientID = getEnvOrDefault("GOOGLE_CLIENT_ID", "1087565480706-8ntgu6rrcbpfmtnlqd2pair903q664v5.apps.googleusercontent.com")
	app.InternalTokenSecret = os.Getenv("INTERNAL_TOKEN_SECRET")
	app.StripeSecretKey = os.Getenv("STRIPE_SECRET_KEY")
	app.StripeAPIBaseURL = getEnvOrDefault("STRIPE_API_BASE_URL", "https://api.stripe.com")
	app.StripeHTTP = &http.Client{Timeout: 15 * time.Second}
	app.PublicBaseURL = strings.TrimRight(os.Getenv("PUBLIC_BASE_URL"), "/")
	app.NotesMaxMemoryBytes = getEnvInt64("NOTES_REDIS_MAX_MEMORY_BYTES", 128*1024*1024)
	app.Analytics = kjv.NewAnalytics(os.Getenv("DATABASE_URL"))
	app.Analytics.Cleanup()
	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			app.Analytics.Cleanup()
		}
	}()
	app.NormalizeAuthConfig()

	// Wait for OpenSearch to be ready before starting
	if app.OpenSearchURL != "" {
		if err := waitForOpenSearch(app.OpenSearchURL, app.OpenSearchHTTP, 30*time.Second); err != nil {
			log.Warnf("OpenSearch not reachable at startup: %v — will retry in background", err)
		} else {
			log.Info("OpenSearch is ready")
		}
	}

	app.SetupRouter()
	app.Analytics.WrapRouter(router)
	router.Handle("/metrics", app.Analytics.MetricsHandler()).Methods("GET")
	app.InitOpenSearch()

	if err := app.PreloadVerses(); err != nil {
		log.Warnf("Initial preload failed: %v — will retry in background every 30s", err)
		go func() {
			for {
				time.Sleep(30 * time.Second)
				if err := app.PreloadVerses(); err == nil {
					log.Infof("Background preload succeeded: %d verses cached", len(app.FlatVerses))
					return
				}
			}
		}()
	} else {
		log.Infof("Preloaded %d verses into memory", len(app.FlatVerses))
	}

	port := ":8000"
	log.Infof("Listening on %s\n", port)
	// Serve
	log.Fatal(http.ListenAndServe(port, removeTrailingSlash(router)))
}

func getEnvOrDefault(key, def string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return def
}

func getEnvInt(key string, def int) int {
	if val := os.Getenv(key); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil {
			return parsed
		}
	}
	return def
}

func getEnvInt64(key string, def int64) int64 {
	if val := os.Getenv(key); val != "" {
		if parsed, err := strconv.ParseInt(val, 10, 64); err == nil && parsed >= 0 {
			return parsed
		}
	}
	return def
}

func parseBoolEnv(value string, def bool) bool {
	if value == "" {
		return def
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return def
	}
	return parsed
}

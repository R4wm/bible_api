package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"github.com/r4wm/bible_api/kjv"
	"github.com/r4wm/bible_api/middleware"
	"github.com/r4wm/mintz5/db"
	"github.com/r4wm/sqlite3_kjv"
	log "github.com/sirupsen/logrus"
)

func removeTrailingSlash(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = strings.TrimSuffix(r.URL.Path, "/")
		next.ServeHTTP(w, r)
	})
}

func main() {

	debug.PrintStack()
	dbPath := flag.String("dbPath", "/tmp/kjv.db", "Path to kjv database.")
	createDB := flag.Bool("createDB", false, "Create the kjv database.")
	flag.Parse()

	// Create the DB if asked
	if *createDB == true {
		path, err := os.Stat(*dbPath)
		if os.IsNotExist(err) {
			_, err := sqlite3_kjv.CreateKJVDB(*dbPath)

			if err != nil {
				panic(err)
			}

			log.Infof("Created database %v", path)
			return // dont run it else docker image build will never finish
		}
	}

	// We didnt create a database, lets go
	// Check the db path exists
	_, err := os.Stat(*dbPath)
	if os.IsNotExist(err) {
		log.Errorf("database path does not exist: %s", *dbPath)
		fmt.Println("Provide dbPath else use createDB argument")
		flag.PrintDefaults()
		os.Exit(1)
	}
	// Create database connection
	database, err := db.CreateDatabase(*dbPath)
	if err != nil {
		panic(err)
	}
	log.Infof("Database connection OK.")

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
		Router:   router,
		Database: database,
		Redis:    rdb,
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
	app.normalizeAuthConfig()
	app.SetupRouter()
	app.InitOpenSearch()
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

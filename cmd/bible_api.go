package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"

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
	// Configure logging
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}
	
	level, err := log.ParseLevel(logLevel)
	if err != nil {
		log.Warnf("Invalid log level '%s', defaulting to info", logLevel)
		level = log.InfoLevel
	}
	log.SetLevel(level)
	
	// Set JSON formatter for structured logging
	log.SetFormatter(&log.JSONFormatter{
		TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
	})
	
	// Configuration from environment variables with sensible defaults
	defaultDBPath := os.Getenv("DB_PATH")
	if defaultDBPath == "" {
		defaultDBPath = "/tmp/kjv.db"
	}
	
	defaultPort := os.Getenv("PORT")
	if defaultPort == "" {
		defaultPort = "8000"
	}

	dbPath := flag.String("dbPath", defaultDBPath, "Path to kjv database.")
	createDB := flag.Bool("createDB", false, "Create the kjv database.")
	port := flag.String("port", defaultPort, "Port to listen on.")
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
	_, err = os.Stat(*dbPath)
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
	app.SetupRouter()
	listenAddr := ":" + *port
	log.Infof("Starting Bible API server on %s", listenAddr)
	// Serve
	log.Fatal(http.ListenAndServe(listenAddr, removeTrailingSlash(router)))
}

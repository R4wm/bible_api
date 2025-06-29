package kjv

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/r4wm/bible_api/middleware"
)

// RateLimitStatus represents the current rate limit status for an IP
type RateLimitStatus struct {
	IP       string `json:"ip"`
	Requests int64  `json:"requests"`
	Blocked  bool   `json:"blocked"`
	Limit    int    `json:"limit"`
	Window   string `json:"window"`
	BlockTTL string `json:"block_ttl"`
}

// BlockIPRequest represents a request to block an IP
type BlockIPRequest struct {
	IP       string `json:"ip"`
	Duration string `json:"duration"` // e.g., "5m", "1h", "24h"
}

// SetupAdminRoutes adds admin endpoints for rate limit management
func (app *App) SetupAdminRoutes() {
	// Admin subrouter (you might want to add authentication here)
	admin := app.Router.PathPrefix("/admin").Subrouter()

	// Get rate limit status for an IP
	admin.HandleFunc("/rate-limit/{ip}", app.getRateLimitStatus).Methods("GET")

	// Block an IP manually
	admin.HandleFunc("/block-ip", app.blockIP).Methods("POST")

	// Unblock an IP
	admin.HandleFunc("/unblock-ip/{ip}", app.unblockIP).Methods("DELETE")

	// Get all blocked IPs (this would require additional Redis tracking)
	admin.HandleFunc("/blocked-ips", app.getBlockedIPs).Methods("GET")

	// Reset all rate limits (development helper)
	admin.HandleFunc("/reset-rate-limits", app.resetAllRateLimits).Methods("POST")
}

func (app *App) getRateLimitStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ip := vars["ip"]

	if ip == "" {
		http.Error(w, "IP address is required", http.StatusBadRequest)
		return
	}

	// Create a rate limiter instance to check status
	rateLimiter := middleware.NewRateLimiter(app.Redis)
	requests, blocked, err := rateLimiter.GetRateLimitStatus(ip)
	if err != nil {
		http.Error(w, "Failed to get rate limit status", http.StatusInternalServerError)
		return
	}

	status := RateLimitStatus{
		IP:       ip,
		Requests: requests,
		Blocked:  blocked,
		Limit:    5,          // This should match your rate limiter config
		Window:   "1 second", // This should match your rate limiter config
		BlockTTL: "1 minute", // This should match your rate limiter config
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

func (app *App) blockIP(w http.ResponseWriter, r *http.Request) {
	var req BlockIPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.IP == "" {
		http.Error(w, "IP address is required", http.StatusBadRequest)
		return
	}

	duration, err := time.ParseDuration(req.Duration)
	if err != nil {
		http.Error(w, "Invalid duration format. Use formats like '5m', '1h', '24h'", http.StatusBadRequest)
		return
	}

	rateLimiter := middleware.NewRateLimiter(app.Redis)
	if err := rateLimiter.BlockIP(req.IP, duration); err != nil {
		log.Printf("Admin: Failed to block IP %s for duration %v: %v", req.IP, duration, err)
		http.Error(w, "Failed to block IP", http.StatusInternalServerError)
		return
	}

	log.Printf("Admin: Manually blocked IP %s for duration %v", req.IP, duration)

	response := map[string]interface{}{
		"message":  "IP blocked successfully",
		"ip":       req.IP,
		"duration": req.Duration,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (app *App) unblockIP(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ip := vars["ip"]

	if ip == "" {
		http.Error(w, "IP address is required", http.StatusBadRequest)
		return
	}

	rateLimiter := middleware.NewRateLimiter(app.Redis)
	if err := rateLimiter.UnblockIP(ip); err != nil {
		log.Printf("Admin: Failed to unblock IP %s: %v", ip, err)
		http.Error(w, "Failed to unblock IP", http.StatusInternalServerError)
		return
	}

	log.Printf("Admin: Manually unblocked IP %s", ip)

	response := map[string]interface{}{
		"message": "IP unblocked successfully",
		"ip":      ip,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (app *App) getBlockedIPs(w http.ResponseWriter, r *http.Request) {
	// This is a simplified implementation
	// In a production environment, you might want to track blocked IPs separately

	response := map[string]interface{}{
		"message": "This endpoint would list all blocked IPs",
		"note":    "Implementation depends on your Redis key management strategy",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (app *App) resetAllRateLimits(w http.ResponseWriter, r *http.Request) {
	// Get Redis client context
	ctx := r.Context()

	// Count keys before deletion for reporting
	rateLimitKeys, err := app.Redis.Keys(ctx, "rate_limit:*").Result()
	if err != nil {
		log.Printf("Admin: Failed to get rate limit keys: %v", err)
		http.Error(w, "Failed to get rate limit keys", http.StatusInternalServerError)
		return
	}

	blockedKeys, err := app.Redis.Keys(ctx, "blocked:*").Result()
	if err != nil {
		log.Printf("Admin: Failed to get blocked keys: %v", err)
		http.Error(w, "Failed to get blocked keys", http.StatusInternalServerError)
		return
	}

	// Delete all rate limit keys
	var deletedRateLimit, deletedBlocked int64
	if len(rateLimitKeys) > 0 {
		deletedRateLimit, err = app.Redis.Del(ctx, rateLimitKeys...).Result()
		if err != nil {
			log.Printf("Admin: Failed to delete rate limit keys: %v", err)
			http.Error(w, "Failed to delete rate limit keys", http.StatusInternalServerError)
			return
		}
	}

	// Delete all blocked keys
	if len(blockedKeys) > 0 {
		deletedBlocked, err = app.Redis.Del(ctx, blockedKeys...).Result()
		if err != nil {
			log.Printf("Admin: Failed to delete blocked keys: %v", err)
			http.Error(w, "Failed to delete blocked keys", http.StatusInternalServerError)
			return
		}
	}

	log.Printf("Admin: Reset all rate limits - deleted %d rate limit keys and %d blocked keys", deletedRateLimit, deletedBlocked)

	response := map[string]interface{}{
		"message":             "All rate limits reset successfully",
		"rate_limit_keys":     deletedRateLimit,
		"blocked_keys":        deletedBlocked,
		"total_keys_deleted":  deletedRateLimit + deletedBlocked,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

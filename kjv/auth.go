package kjv

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"google.golang.org/api/idtoken"
)

const (
	defaultJWTIssuer       = "bible_api"
	defaultJWTAudience     = "bible_api_clients"
	defaultJWTTTLSeconds   = 3600
	defaultSessionTTL      = 3600
	defaultSessionCookie   = "bible_api_session"
	defaultSynonymsScope   = "synonyms:write"
	defaultTokenTTLSeconds = 900
)

type AppClaims struct {
	Scope string `json:"scope"`
	jwt.RegisteredClaims
}

type sessionData struct {
	Token string `json:"token"`
	Sub   string `json:"sub"`
	Scope string `json:"scope"`
	Exp   int64  `json:"exp"`
}

func (app *App) SetupAuthRoutes() {
	app.Router.HandleFunc("/auth/google/token", app.googleToken).Methods("POST")
	app.Router.HandleFunc("/auth/config", app.authConfig).Methods("GET")
	app.Router.HandleFunc("/auth/me", app.authMe).Methods("GET")
	app.Router.HandleFunc("/auth/logout", app.authLogout).Methods("POST")

	admin := app.Router.PathPrefix("/admin").Subrouter()
	admin.HandleFunc("/token", app.internalToken).Methods("POST")
}

func (app *App) authConfig(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"google_client_id": app.GoogleClientID,
	})
}

func (app *App) googleToken(w http.ResponseWriter, r *http.Request) {
	if app.GoogleClientID == "" {
		jsonError(w, http.StatusServiceUnavailable, "Google auth not configured")
		return
	}
	var req struct {
		IDToken    string `json:"id_token"`
		Scope      string `json:"scope"`
		TTLSeconds int    `json:"ttl_seconds"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if strings.TrimSpace(req.IDToken) == "" {
		jsonError(w, http.StatusBadRequest, "id_token is required")
		return
	}

	payload, err := idtoken.Validate(context.Background(), req.IDToken, app.GoogleClientID)
	if err != nil {
		jsonError(w, http.StatusUnauthorized, "invalid google token")
		return
	}

	scope := strings.TrimSpace(req.Scope)
	if scope == "" {
		scope = defaultSynonymsScope
	}

	ttl := clampTTL(req.TTLSeconds, app.JWTTTLSeconds)
	token, exp, err := app.mintJWT(payload.Subject, scope, ttl)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "failed to mint token")
		return
	}

	if err := app.createSession(w, sessionData{
		Token: token,
		Sub:   payload.Subject,
		Scope: scope,
		Exp:   exp,
	}); err != nil {
		jsonError(w, http.StatusInternalServerError, "failed to create session")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"token":      token,
		"expires_in": ttl,
	})
}

func (app *App) internalToken(w http.ResponseWriter, r *http.Request) {
	if app.InternalTokenSecret == "" {
		jsonError(w, http.StatusServiceUnavailable, "internal token endpoint not configured")
		return
	}
	if r.Header.Get("X-Internal-Secret") != app.InternalTokenSecret {
		jsonError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req struct {
		Sub        string `json:"sub"`
		Scope      string `json:"scope"`
		TTLSeconds int    `json:"ttl_seconds"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	sub := strings.TrimSpace(req.Sub)
	if sub == "" {
		sub = "internal"
	}
	scope := strings.TrimSpace(req.Scope)
	if scope == "" {
		scope = defaultSynonymsScope
	}

	ttl := clampTTL(req.TTLSeconds, app.JWTTTLSeconds)
	token, exp, err := app.mintJWT(sub, scope, ttl)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, "failed to mint token")
		return
	}
	if err := app.createSession(w, sessionData{
		Token: token,
		Sub:   sub,
		Scope: scope,
		Exp:   exp,
	}); err != nil {
		jsonError(w, http.StatusInternalServerError, "failed to create session")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"token":      token,
		"expires_in": ttl,
	})
}

func (app *App) authMe(w http.ResponseWriter, r *http.Request) {
	session, err := app.getSession(r)
	if err != nil {
		jsonError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"token": session.Token,
		"sub":   session.Sub,
		"scope": session.Scope,
		"exp":   session.Exp,
	})
}

func (app *App) authLogout(w http.ResponseWriter, r *http.Request) {
	_ = app.deleteSession(w, r)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status": "ok",
	})
}

func (app *App) mintJWT(sub, scope string, ttl int) (string, int64, error) {
	if len(app.JWTSecret) == 0 {
		return "", 0, errors.New("JWT secret not configured")
	}
	now := time.Now()
	exp := now.Add(time.Duration(ttl) * time.Second).Unix()
	claims := AppClaims{
		Scope: scope,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   sub,
			Issuer:    app.JWTIssuer,
			Audience:  []string{app.JWTAudience},
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(time.Unix(exp, 0)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(app.JWTSecret)
	if err != nil {
		return "", 0, err
	}
	return signed, exp, nil
}

func (app *App) jwtMiddleware(requiredScope string) muxMiddleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if len(app.JWTSecret) == 0 {
				jsonError(w, http.StatusServiceUnavailable, "JWT not configured")
				return
			}
			authHeader := r.Header.Get("Authorization")
			if !strings.HasPrefix(authHeader, "Bearer ") {
				jsonError(w, http.StatusUnauthorized, "missing bearer token")
				return
			}
			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
			claims := &AppClaims{}
			parsed, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
				return app.JWTSecret, nil
			})
			if err != nil || !parsed.Valid {
				jsonError(w, http.StatusUnauthorized, "invalid token")
				return
			}
			if !claims.VerifyIssuer(app.JWTIssuer, true) || !claims.VerifyAudience(app.JWTAudience, true) {
				jsonError(w, http.StatusUnauthorized, "invalid token claims")
				return
			}
			if !hasScope(claims.Scope, requiredScope) {
				jsonError(w, http.StatusForbidden, "insufficient scope")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

type muxMiddleware func(http.Handler) http.Handler

func hasScope(scope, required string) bool {
	for _, part := range strings.Fields(scope) {
		if part == required {
			return true
		}
	}
	return false
}

func clampTTL(ttl, def int) int {
	if ttl <= 0 {
		return def
	}
	if ttl > 86400 {
		return 86400
	}
	return ttl
}

func (app *App) createSession(w http.ResponseWriter, data sessionData) error {
	id, err := randomToken(32)
	if err != nil {
		return err
	}
	key := "session:" + id
	payload, _ := json.Marshal(data)
	ctx := context.Background()
	ttl := time.Duration(app.SessionTTLSeconds) * time.Second
	if app.SessionTTLSeconds <= 0 {
		ttl = time.Duration(defaultSessionTTL) * time.Second
	}
	if err := app.Redis.Set(ctx, key, payload, ttl).Err(); err != nil {
		return err
	}
	cookie := &http.Cookie{
		Name:     app.SessionCookieName,
		Value:    id,
		Path:     "/",
		HttpOnly: true,
		Secure:   app.SessionCookieSecure,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, cookie)
	return nil
}

func (app *App) getSession(r *http.Request) (*sessionData, error) {
	cookie, err := r.Cookie(app.SessionCookieName)
	if err != nil {
		return nil, err
	}
	key := "session:" + cookie.Value
	ctx := context.Background()
	val, err := app.Redis.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	var data sessionData
	if err := json.Unmarshal([]byte(val), &data); err != nil {
		return nil, err
	}
	return &data, nil
}

func (app *App) deleteSession(w http.ResponseWriter, r *http.Request) error {
	cookie, err := r.Cookie(app.SessionCookieName)
	if err != nil {
		return err
	}
	key := "session:" + cookie.Value
	ctx := context.Background()
	_ = app.Redis.Del(ctx, key).Err()
	expired := &http.Cookie{
		Name:     app.SessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   app.SessionCookieSecure,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, expired)
	return nil
}

func randomToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func (app *App) normalizeAuthConfig() {
	if app.JWTIssuer == "" {
		app.JWTIssuer = defaultJWTIssuer
	}
	if app.JWTAudience == "" {
		app.JWTAudience = defaultJWTAudience
	}
	if app.JWTTTLSeconds == 0 {
		app.JWTTTLSeconds = defaultJWTTTLSeconds
	}
	if app.SessionTTLSeconds == 0 {
		app.SessionTTLSeconds = defaultSessionTTL
	}
	if app.SessionCookieName == "" {
		app.SessionCookieName = defaultSessionCookie
	}
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

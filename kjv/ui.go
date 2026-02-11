package kjv

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const uiDistDir = "web/dist"

func (app *App) SetupUIRoutes() {
	handler := app.uiStaticHandler()
	app.Router.PathPrefix("/v2/").Handler(handler)
	app.Router.Handle("/v2", handler)
}

func (app *App) uiStaticHandler() http.Handler {
	distDir := resolveUIDistDir()
	fileServer := http.FileServer(http.Dir(distDir))
	indexPath := filepath.Join(distDir, "index.html")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v2" || r.URL.Path == "/v2/" {
			http.ServeFile(w, r, indexPath)
			return
		}

		rel := strings.TrimPrefix(r.URL.Path, "/v2/")
		if rel == "" {
			http.ServeFile(w, r, indexPath)
			return
		}

		fsPath := filepath.Join(distDir, rel)
		if info, err := os.Stat(fsPath); err == nil && !info.IsDir() {
			r.URL.Path = "/" + rel
			fileServer.ServeHTTP(w, r)
			return
		}

		http.ServeFile(w, r, indexPath)
	})
}

func resolveUIDistDir() string {
	if env := os.Getenv("UI_DIST_DIR"); env != "" {
		return env
	}

	candidates := []string{uiDistDir}
	if wd, err := os.Getwd(); err == nil {
		candidates = append(candidates, filepath.Join(wd, uiDistDir))
	}
	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		candidates = append(candidates, filepath.Join(exeDir, uiDistDir))
		candidates = append(candidates, filepath.Join(exeDir, "..", uiDistDir))
		candidates = append(candidates, filepath.Join("/go/src/bible_api", uiDistDir))
	}

	for _, candidate := range candidates {
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate
		}
	}
	return uiDistDir
}

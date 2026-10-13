package web

import (
	"encoding/json"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/andrearcaina/draftkings-odds/internal/draftkings"
	"github.com/andrearcaina/draftkings-odds/internal/logger"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Handler struct {
	odds            *draftkings.Client
	logger          *logger.Logger
	refreshInterval time.Duration
}

func NewHandler(odds *draftkings.Client, logger *logger.Logger, refreshInterval time.Duration) *Handler {
	return &Handler{
		odds:            odds,
		logger:          logger,
		refreshInterval: refreshInterval,
	}
}

func (h *Handler) Routes() chi.Router {
	router := chi.NewRouter()

	router.Use(middleware.Recoverer, h.logger.RequestLogger, securityHeaders)
	router.Get("/", h.index)
	router.Get("/static/index.js", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/javascript; charset=utf-8")
		_, _ = w.Write(indexJS)
	})
	router.Get("/static/index.css", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
		_, _ = w.Write(indexCSS)
	})
	router.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.FS(staticAssets))))
	router.Get("/api/odds", h.oddsJSON)
	router.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok\n"))
	})

	return router
}

func (h *Handler) index(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Cache-Control", "no-store")

	page := strings.Replace(indexHTML, "__REFRESH_INTERVAL_MS__", strconv.FormatInt(h.refreshInterval.Milliseconds(), 10), 1)
	page = strings.Replace(page, "__APP_JS__", string(indexJS), 1)
	template.Must(template.New("index").Parse(page)).Execute(w, nil)
}

func (h *Handler) oddsJSON(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Get("refresh") == "1" {
		h.odds.Refresh(r.Context())
	}

	w.Header().Set("Content-Type", "application/json")

	_ = json.NewEncoder(w).Encode(h.odds.Snapshot())
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval' https://cdn.jsdelivr.net; style-src 'self' 'unsafe-inline'; connect-src 'self'; font-src 'self' data:; frame-ancestors 'none'")

		next.ServeHTTP(w, r)
	})
}

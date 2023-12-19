package topi

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"strings"

	"github.com/alecthomas/chroma/v2"
	chtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/stampede"
	"github.com/topi314/slog-chi"
)

var (
	StyleDark  = styles.Get("github-dark")
	StyleLight = styles.Get("github")
)

func (s *Server) Routes() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.CleanPath)
	r.Use(middleware.RealIP)
	r.Use(middleware.RequestID)
	r.Use(middleware.Compress(5))
	r.Use(slogchi.NewWithConfig(slog.Default(), slogchi.Config{
		DefaultLevel:     slog.LevelInfo,
		ClientErrorLevel: slog.LevelDebug,
		ServerErrorLevel: slog.LevelError,
		WithRequestID:    true,
		Filters: []slogchi.Filter{
			slogchi.IgnorePathPrefix("/assets"),
		},
	}))
	r.Use(cacheControl)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Heartbeat("/ping"))

	r.Mount("/assets", http.FileServer(s.assets))
	r.Get("/dark.css", s.theme(StyleDark))
	r.Get("/light.css", s.theme(StyleLight))
	r.Handle("/robots.txt", s.file("/assets/robots.txt"))

	stampedeMiddleware := func(handler http.Handler) http.Handler { return handler }
	lastFMStampedeMiddleware := func(handler http.Handler) http.Handler { return handler }
	if s.cfg.Cache != nil && s.cfg.Cache.Size > 0 && s.cfg.Cache.TTL > 0 {
		stampedeMiddleware = stampede.HandlerWithKey(s.cfg.Cache.Size, s.cfg.Cache.TTL, cacheKeyFunc)
	}
	if s.cfg.LastFM.Size > 0 && s.cfg.LastFM.TTL > 0 {
		lastFMStampedeMiddleware = stampede.HandlerWithKey(s.cfg.LastFM.Size, s.cfg.LastFM.TTL, cacheKeyFunc)
	}

	r.Group(func(r chi.Router) {
		r.Route("/api", func(r chi.Router) {
			r.Route("/repositories", func(r chi.Router) {
				r.Use(stampedeMiddleware)
				r.Get("/", s.repositories)
			})
			r.Route("/lastfm", func(r chi.Router) {
				r.Use(lastFMStampedeMiddleware)
				r.Get("/", s.lastfm)
			})
		})
		r.Route("/", func(r chi.Router) {
			r.Use(stampedeMiddleware)
			r.Get("/", s.index)
			r.Head("/", s.index)
		})
	})
	r.NotFound(s.redirectRoot)
	return r
}

func cacheKeyFunc(r *http.Request) uint64 {
	theme := "dark"
	cookie, _ := r.Cookie("theme")
	if cookie != nil {
		theme = cookie.Value
	}
	return stampede.BytesToHash([]byte(theme), []byte(strings.ToLower(r.URL.Path)), []byte(r.URL.RawQuery))
}

func (s *Server) repositories(w http.ResponseWriter, r *http.Request) {
	after := r.URL.Query().Get("after")
	ctx := r.Context()
	vars, err := s.FetchRepositories(ctx, after)
	if err != nil {
		slog.ErrorContext(ctx, "failed to fetch repositories", slog.Any("error", err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err = s.tmpl(w, "projects.gohtml", vars); err != nil {
		slog.ErrorContext(ctx, "failed to render projects template", slog.Any("error", err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (s *Server) lastfm(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := s.FetchLastFM(ctx)

	if err := s.tmpl(w, "lastfm.gohtml", vars); err != nil {
		slog.ErrorContext(ctx, "failed to render lastfm template", slog.Any("error", err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (s *Server) index(w http.ResponseWriter, r *http.Request) {
	vars, err := s.FetchData(r.Context())
	if err != nil {
		s.error(w, r, fmt.Errorf("failed to fetch data: %w", err), http.StatusInternalServerError)
		return
	}

	if themeCookie, _ := r.Cookie("theme"); themeCookie != nil {
		vars.Dark = themeCookie.Value == "dark"
	}

	if err = s.HighlightData(vars); err != nil {
		s.error(w, r, fmt.Errorf("failed to highlight data: %w", err), http.StatusInternalServerError)
		return
	}

	if err = s.tmpl(w, "index.gohtml", vars); err != nil {
		log.Println("failed to execute template:", err)
	}
}

func (s *Server) theme(style *chroma.Style) http.HandlerFunc {
	cssBuff := new(bytes.Buffer)
	if err := chtml.New(chtml.WithClasses(true), chtml.ClassPrefix("ch-")).WriteCSS(cssBuff, style); err != nil {
		return func(w http.ResponseWriter, r *http.Request) {
			s.error(w, r, fmt.Errorf("failed to write CSS: %w", err), http.StatusInternalServerError)
		}
	}

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css")
		_, _ = w.Write(cssBuff.Bytes())
	}
}

func (s *Server) redirectRoot(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (s *Server) error(w http.ResponseWriter, r *http.Request, err error, status int) {
	if status == http.StatusInternalServerError {
		slog.ErrorContext(r.Context(), "internal server error", slog.Any("error", err))
	}
	w.WriteHeader(status)

	vars := map[string]any{
		"Error":     err.Error(),
		"Status":    status,
		"RequestID": middleware.GetReqID(r.Context()),
		"Path":      r.URL.Path,
	}
	if tmplErr := s.tmpl(w, "error.gohtml", vars); tmplErr != nil && !errors.Is(tmplErr, http.ErrHandlerTimeout) {
		slog.ErrorContext(r.Context(), "failed to render error template", slog.Any("error", tmplErr))
	}
}

func (s *Server) file(path string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		file, err := s.assets.Open(path)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer file.Close()
		_, _ = io.Copy(w, file)
	}
}

func cacheControl(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/assets/") {
			w.Header().Set("Cache-Control", "public, max-age=86400")
			next.ServeHTTP(w, r)
			return
		}
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		next.ServeHTTP(w, r)
	})
}

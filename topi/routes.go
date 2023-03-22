package topi

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/alecthomas/chroma/v2/styles"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/stampede"
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
	r.Use(middleware.Maybe(
		middleware.RequestLogger(&middleware.DefaultLogFormatter{
			Logger: log.Default(),
		}),
		func(r *http.Request) bool {
			// Don't log requests for assets
			return !strings.HasPrefix(r.URL.Path, "/assets")
		},
	))
	r.Use(middleware.Recoverer)
	r.Use(middleware.Heartbeat("/ping"))

	r.Mount("/assets", http.FileServer(s.assets))
	r.Handle("/favicon.ico", s.file("/assets/favicon.png"))
	r.Handle("/favicon.png", s.file("/assets/favicon.png"))
	r.Handle("/favicon-light.png", s.file("/assets/favicon-light.png"))
	r.Handle("/robots.txt", s.file("/assets/robots.txt"))

	r.Group(func(r chi.Router) {
		if s.cfg.Cache != nil && s.cfg.Cache.Size > 0 {
			r.Use(stampede.HandlerWithKey(s.cfg.Cache.Size, s.cfg.Cache.TTL, cacheKeyFunc))
		}
		r.Get("/", s.index)
		r.Head("/", s.index)
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
	return stampede.BytesToHash([]byte(theme), []byte(strings.ToLower(r.URL.Path)))
}

func (s *Server) index(w http.ResponseWriter, r *http.Request) {
	vars, err := s.FetchData(r.Context())
	if err != nil {
		s.prettyError(w, r, fmt.Errorf("failed to fetch data: %w", err), http.StatusInternalServerError)
		return
	}

	style := StyleDark
	themeCookie, _ := r.Cookie("theme")
	if themeCookie != nil {
		vars.Dark = themeCookie.Value == "dark"
		if !vars.Dark {
			style = StyleLight
		}
	}

	if err = s.HighlightData(vars, style); err != nil {
		s.prettyError(w, r, fmt.Errorf("failed to highlight data: %w", err), http.StatusInternalServerError)
		return
	}

	if err = s.tmpl(w, "index.gohtml", vars); err != nil {
		log.Println("failed to execute template:", err)
	}
}

func (s *Server) redirectRoot(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (s *Server) log(r *http.Request, logType string, err error) {
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return
	}
	log.Printf("Error while handling %s(%s) %s: %s\n", logType, middleware.GetReqID(r.Context()), r.RequestURI, err)
}

func (s *Server) prettyError(w http.ResponseWriter, r *http.Request, err error, status int) {
	if status == http.StatusInternalServerError {
		s.log(r, "pretty request", err)
	}
	w.WriteHeader(status)

	vars := map[string]any{
		"Error":     err.Error(),
		"Status":    status,
		"RequestID": middleware.GetReqID(r.Context()),
		"Path":      r.URL.Path,
	}
	if tmplErr := s.tmpl(w, "error.gohtml", vars); tmplErr != nil && tmplErr != http.ErrHandlerTimeout {
		s.log(r, "template", tmplErr)
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

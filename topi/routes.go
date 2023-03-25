package topi

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	chtml "github.com/alecthomas/chroma/v2/formatters/html"
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
	r.HandleFunc("/dark.css", s.theme(true))
	r.HandleFunc("/light.css", s.theme(false))
	r.Handle("/favicon.ico", s.file("/assets/favicon.png"))
	r.Handle("/favicon.png", s.file("/assets/favicon.png"))
	r.Handle("/favicon-light.png", s.file("/assets/favicon-light.png"))
	r.Handle("/robots.txt", s.file("/assets/robots.txt"))

	r.Group(func(r chi.Router) {
		if s.cfg.Cache != nil && s.cfg.Cache.Size > 0 {
			r.Use(stampede.HandlerWithKey(s.cfg.Cache.Size, s.cfg.Cache.TTL, cacheKeyFunc))
		}
		r.Route("/api", func(r chi.Router) {
			r.Route("/posts", func(r chi.Router) {
				r.Get("/", s.posts)
				r.Head("/", s.posts)
			})
			r.Route("/repositories", func(r chi.Router) {
				r.Get("/", s.repositories)
				r.Head("/", s.repositories)
			})
		})
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

func (s *Server) repositories(w http.ResponseWriter, r *http.Request) {
	after := r.URL.Query().Get("after")
	vars, err := s.FetchRepositories(r.Context(), after)
	if err != nil {
		s.log(r, "api request", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err = s.tmpl(w, "projects.gohtml", vars); err != nil {
		s.log(r, "template", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (s *Server) posts(w http.ResponseWriter, r *http.Request) {
	after := r.URL.Query().Get("after")
	vars, err := s.FetchPosts(r.Context(), after)
	if err != nil {
		s.log(r, "api request", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err = s.tmpl(w, "blog.gohtml", vars); err != nil {
		s.log(r, "template", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (s *Server) index(w http.ResponseWriter, r *http.Request) {
	vars, err := s.FetchData(r.Context())
	if err != nil {
		s.prettyError(w, r, fmt.Errorf("failed to fetch data: %w", err), http.StatusInternalServerError)
		return
	}

	if themeCookie, _ := r.Cookie("theme"); themeCookie != nil {
		vars.Dark = themeCookie.Value == "dark"
	}

	if err = s.HighlightData(vars); err != nil {
		s.prettyError(w, r, fmt.Errorf("failed to highlight data: %w", err), http.StatusInternalServerError)
		return
	}

	if err = s.tmpl(w, "index.gohtml", vars); err != nil {
		log.Println("failed to execute template:", err)
	}
}

func (s *Server) theme(dark bool) http.HandlerFunc {
	style := StyleDark
	if !dark {
		style = StyleLight
	}
	cssBuff := new(bytes.Buffer)
	if err := chtml.New(chtml.WithClasses(true), chtml.ClassPrefix("ch-")).WriteCSS(cssBuff, style); err != nil {
		return func(w http.ResponseWriter, r *http.Request) {
			s.prettyError(w, r, fmt.Errorf("failed to write CSS: %w", err), http.StatusInternalServerError)
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

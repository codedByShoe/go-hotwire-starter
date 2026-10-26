package web

import (
	"net/http"
	"strings"
	"time"

	"github.com/codedbyshoe/go-hotwire-starter/internal/db/g"
	"github.com/codedbyshoe/go-hotwire-starter/internal/models"
	"github.com/codedbyshoe/go-hotwire-starter/pages"
	"gorm.io/gorm"
)

func newServer(importMap string, db *gorm.DB) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("GET /assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))
	mux.Handle("GET /configurations/", http.StripPrefix("/configurations/", http.FileServer(http.Dir("configurations"))))

	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		render(w, r, pages.IndexPage(nativeClass(r), importMap))
	})
	mux.HandleFunc("GET /about", func(w http.ResponseWriter, r *http.Request) {
		render(w, r, pages.AboutPage(nativeClass(r), importMap))
	})
	mux.HandleFunc("GET /time", func(w http.ResponseWriter, r *http.Request) {
		render(w, r, pages.TimeFrame(time.Now().Format(time.RFC1123)))
	})

	mux.HandleFunc("GET /notes", func(w http.ResponseWriter, r *http.Request) {
		notes, err := g.Notes[models.Note](db).Recent(r.Context(), 50)
		if err != nil {
			serverError(w, r, err)
			return
		}
		render(w, r, pages.NotesPage(nativeClass(r), importMap, notes, ""))
	})
	mux.HandleFunc("POST /notes", func(w http.ResponseWriter, r *http.Request) {
		body := strings.TrimSpace(r.FormValue("body"))
		if body == "" {
			notes, err := g.Notes[models.Note](db).Recent(r.Context(), 50)
			if err != nil {
				serverError(w, r, err)
				return
			}
			// Turbo renders 4xx form responses in place.
			w.WriteHeader(http.StatusUnprocessableEntity)
			render(w, r, pages.NotesPage(nativeClass(r), importMap, notes, "Write something first."))
			return
		}
		if err := gorm.G[models.Note](db).Create(r.Context(), &models.Note{Body: body}); err != nil {
			serverError(w, r, err)
			return
		}
		// Turbo expects a 303 after a successful form submission.
		http.Redirect(w, r, "/notes", http.StatusSeeOther)
	})

	mux.HandleFunc("GET /up", func(w http.ResponseWriter, r *http.Request) {
		sqlDB, err := db.DB()
		if err == nil {
			err = sqlDB.PingContext(r.Context())
		}
		if err != nil {
			http.Error(w, "database unavailable", http.StatusServiceUnavailable)
			return
		}
		w.Write([]byte("ok"))
	})

	return recoverPanics(logRequests(mux))
}

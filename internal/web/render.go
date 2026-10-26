package web

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/a-h/templ"
)

func render(w http.ResponseWriter, r *http.Request, component templ.Component) {
	if w.Header().Get("Content-Type") == "" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
	}
	if err := component.Render(r.Context(), w); err != nil {
		slog.Error("render failed", "path", r.URL.Path, "error", err)
	}
}

func serverError(w http.ResponseWriter, r *http.Request, err error) {
	slog.Error("request failed", "method", r.Method, "path", r.URL.Path, "error", err)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
}

// nativeClass marks pages loaded by a Hotwire Native app so CSS can hide the
// web header and leave navigation chrome to the app.
func nativeClass(r *http.Request) string {
	ua := r.UserAgent()
	if strings.Contains(ua, "Hotwire Native Android;") || strings.Contains(ua, "Hotwire Native iOS;") {
		return "hotwire-native"
	}
	return ""
}

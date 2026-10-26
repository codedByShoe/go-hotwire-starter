package main

import (
	"log/slog"
	"os"

	"github.com/codedbyshoe/go-hotwire-starter/internal/web"
)

func main() {
	if err := web.Run(); err != nil {
		slog.Error("server stopped", "error", err)
		os.Exit(1)
	}
}

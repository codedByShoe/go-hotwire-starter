package web

import (
	"cmp"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/codedbyshoe/go-hotwire-starter/internal/db"
)

func Run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	addr := ":" + cmp.Or(os.Getenv("PORT"), "3000")
	dbPath := cmp.Or(os.Getenv("DATABASE_PATH"), "data/app.db")

	importMap, err := hotwireImportMap(ctx)
	if err != nil {
		return err
	}
	database, err := db.Open(dbPath)
	if err != nil {
		return err
	}
	if err := db.Migrate(database); err != nil {
		return err
	}

	server := &http.Server{
		Addr:              addr,
		Handler:           newServer(importMap, database),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       2 * time.Minute,
	}

	errs := make(chan error, 1)
	go func() {
		slog.Info("listening", "addr", addr, "database", dbPath)
		errs <- server.ListenAndServe()
	}()

	select {
	case err := <-errs:
		return err
	case <-ctx.Done():
	}

	slog.Info("shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return db.Close(database)
}

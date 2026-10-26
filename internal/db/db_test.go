package db

import (
	"context"
	"path/filepath"
	"sync"
	"testing"

	"github.com/codedbyshoe/go-hotwire-starter/internal/db/g"
	"github.com/codedbyshoe/go-hotwire-starter/internal/models"
	"gorm.io/gorm"
)

func openTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	database, err := Open(filepath.Join(t.TempDir(), "nested", "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	if err := Migrate(database); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { Close(database) })
	return database
}

func TestPragmas(t *testing.T) {
	database := openTestDB(t)
	for pragma, want := range map[string]string{
		"journal_mode": "wal",
		"foreign_keys": "1",
		"busy_timeout": "5000",
		"synchronous":  "1", // NORMAL
	} {
		var got string
		if err := database.Raw("PRAGMA " + pragma).Scan(&got).Error; err != nil {
			t.Fatal(err)
		}
		if got != want {
			t.Errorf("PRAGMA %s = %q, want %q", pragma, got, want)
		}
	}
}

func TestNotesRecent(t *testing.T) {
	ctx := context.Background()
	database := openTestDB(t)

	for _, body := range []string{"first", "second", "third"} {
		if err := gorm.G[models.Note](database).Create(ctx, &models.Note{Body: body}); err != nil {
			t.Fatal(err)
		}
	}

	notes, err := g.Notes[models.Note](database).Recent(ctx, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(notes) != 2 || notes[0].Body != "third" || notes[1].Body != "second" {
		t.Fatalf("Recent(2) = %+v", notes)
	}

	count, err := gorm.G[models.Note](database).Where(g.Note.Body.Eq("first")).Count(ctx, "*")
	if err != nil || count != 1 {
		t.Fatalf("count = %d, err = %v", count, err)
	}
}

// Concurrent writers wait for the lock (busy_timeout + _txlock=immediate)
// instead of failing with "database is locked".
func TestConcurrentWrites(t *testing.T) {
	ctx := context.Background()
	database := openTestDB(t)

	var wg sync.WaitGroup
	errs := make(chan error, 50)
	for range 50 {
		wg.Go(func() {
			errs <- gorm.G[models.Note](database).Create(ctx, &models.Note{Body: "x"})
		})
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}

	count, err := gorm.G[models.Note](database).Count(ctx, "*")
	if err != nil || count != 50 {
		t.Fatalf("count = %d, err = %v", count, err)
	}
}

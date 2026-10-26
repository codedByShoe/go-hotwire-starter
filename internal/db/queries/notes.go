// Package queries holds GORM CLI query interfaces. Run `make gen` after
// editing; the implementations are generated into internal/db/g.
package queries

type Notes[T any] interface {
	// SELECT * FROM @@table ORDER BY created_at DESC, id DESC LIMIT @limit
	Recent(limit int) ([]T, error)
}

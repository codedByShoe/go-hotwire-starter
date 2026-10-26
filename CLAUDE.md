# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

This repository is a reusable starter for Go + Hotwire web apps. New projects are created with `gonew` plus `go run ./cmd/rename` (see the README), so keep it generic. The placeholders "Coded by Shoe", `CodedByShoe` and `com.codedbyshoe.hotwire` are what `cmd/rename` replaces. Use them consistently when you add files that need the app name.

## Commands

Requires Go 1.27.1 and a C compiler (CGO is needed by the SQLite driver). Dev tools (templ, GORM CLI, Air) are pinned with `tool` directives in `go.mod`; run them through `go tool`, not global installs. Run everything from the project root, since the server serves `./assets` and `./configurations` relative to the working directory.

```sh
make dev                 # go tool air — live reload, proxies :3000 on :7331
make gen                 # go tool templ generate + go generate ./internal/db (GORM CLI)
make test                # gen, go vet, go test
go run ./cmd/server      # serve on :3000 (PORT, DATABASE_PATH env vars; DB defaults to data/app.db)
go test ./internal/web -run TestHotwireRoutes/about   # single test / subtest
go test ./internal/db -run TestNotesRecent    # DB tests use temp-file databases
```

iOS UI test (server must be running on :3000):

```sh
xcodebuild -project ios/CodedByShoe.xcodeproj -scheme CodedByShoe \
  -destination 'platform=iOS Simulator,name=iPhone 17' CODE_SIGNING_ALLOWED=NO test
```

## Architecture

Stdlib `net/http` renders templ components as HTML. The same URLs are consumed by browsers (via Turbo) and by Hotwire Native mobile apps. There is no Node.js or JS build step. Prefer plain, easy-to-read code over clever abstractions.

- **Server**: `cmd/server/main.go` only calls `web.Run()`. In `internal/web`, `server.go` (`Run`) loads config, opens and migrates the DB, and runs `http.Server` with graceful shutdown. `routes.go` has `newServer(importMap, db)`, which builds the `ServeMux` (Go 1.22 method/path patterns) and wraps it in `middleware.go` (request logging, panic recovery). `render.go` holds the render/error helpers and `importmap.go` the JS pins. Tests in `internal/web/web_test.go` drive `newServer` via `httptest`, after `t.Chdir("../..")` so the relative `assets/` paths resolve. `GET /up` pings the DB and is the deploy healthcheck.
- **templ**: `components/base.templ` is the layout; `pages/*.templ` are pages. Every full page takes `(nativeClass, importMap)` and passes them to `components.Base`. Edit `.templ` sources, never the generated `*_templ.go`.
- **Forms with Turbo**: on success, respond with `303 See Other`. On validation failure, re-render with a 4xx status (`422`), which Turbo renders in place. See `POST /notes`.
- **Native detection**: `nativeClass()` returns `hotwire-native` for Hotwire Native Android/iOS user agents. CSS uses it to hide the web header.
- **Database** (`internal/db`): GORM with `gorm.io/driver/sqlite` (mattn/go-sqlite3, CGO). `Open` uses one connection pool, with the SQLite settings (WAL, busy_timeout, foreign keys, synchronous=NORMAL, `_txlock=immediate`) in the DSN. `Close` closes it. Schema comes from `db.Migrate` (`AutoMigrate`); register new models there.
- **GORM CLI codegen**: models live in `internal/models`, SQL-template query interfaces in `internal/db/queries`. The `//go:generate` lines in `internal/db/db.go` write both into `internal/db/g` (committed). Use `g.<Query>[T](db).Method(ctx, …)` for template queries and `gorm.G[T](db).Where(g.<Model>.<Field>.Eq(…))` for field helpers. Generation doesn't remove stale files, so delete files from `internal/db/g` by hand when you remove a model or interface.
- **JavaScript via import map** (`internal/web/importmap.go`): Turbo, Stimulus and Hotwire Native Bridge are pinned with `go-importmap`. At startup `CacheOrFetch` restores `assets/vendor/` from the versioned `.importmap/` cache (or the CDN if the cache is missing) and the rendered import map is injected into the layout head. To bump a library, change its version/URL there.
- **Stimulus**: controllers are registered manually in `assets/js/app.js`.
- **Bridge components**: a web `BridgeComponent` pairs with a native component of the same name: `button` in `assets/js/controllers/bridge/button_controller.js` ↔ `ios/CodedByShoe/ButtonComponent.swift`, registered in `AppDelegate`. Android has no wrapper app in this repo.
- **Path configurations**: `configurations/{android,ios}_v1.json` are served by the app, and the iOS app bundles `ios_v1.json` as its fallback. Android rules must map to `hotwire://fragment/web` (asserted in tests).
- **Deploy**: `Dockerfile` (CGO build with gcc/musl-dev, Alpine runtime, non-root, `DATABASE_PATH=/data/app.db`) targets Coolify. `docker-entrypoint.sh` runs the server under `litestream replicate -exec` when `LITESTREAM_REPLICA_URL` is set, after restoring onto an empty volume. The server needs `assets/`, `configurations/` and `.importmap/` next to it at runtime.

## Velle-managed CSS/JS

`assets/css/velle/` and `assets/js/velle/` are managed by the `velle` CLI (config in `velle.json`). Do not edit between the `velle:begin` / `velle:end` markers in `main.css` / `main.js`. Site styles go in `assets/css/app.css`, using velle tokens (`var(--color-*)`). Add components with `velle add <name>` rather than hand-copying. Adapter components (chart, datepicker, editor, inputmask, sortable) import a third-party library by bare name; pin that library in `internal/web/importmap.go` under the same name.

## Tests

`internal/web/web_test.go` asserts on rendered HTML substrings (titles, import map entries, frame IDs, controller attributes) and on the path configuration structure. Update the expected strings when you change page copy, titles or vendored library paths.

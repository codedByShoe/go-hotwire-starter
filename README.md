# Go + Hotwire starter

This starter serves HTML from Go's `net/http` and [templ](https://templ.guide), with SQLite through [GORM](https://gorm.io) and styles from [velle-css](https://github.com/codedbyshoe/velle-css). [Turbo](https://turbo.hotwired.dev/) handles web navigation and the live frame on the home page; [Stimulus](https://stimulus.hotwired.dev/) powers its refresh button. The [Hotwire Native Bridge](https://native.hotwired.dev/reference/bridge-installation) is installed for optional native toolbar controls.

## Start a new project

The starter is a Go module, so [`gonew`](https://pkg.go.dev/golang.org/x/tools/cmd/gonew) can copy it and rewrite the module path. gonew only rewrites `.go` files, so the rename tool then fixes the module path in the `.templ` files. It also replaces the display name, Swift identifier and iOS bundle ID, regenerates templ code, and deletes itself:

```sh
go install golang.org/x/tools/cmd/gonew@latest
gonew github.com/codedbyshoe/go-hotwire-starter github.com/you/myapp
cd myapp
go run ./cmd/rename -name "My App" -bundle com.example.myapp
git init && go test ./...
```

Without gonew (for example on a plain `git clone`), pass the module path yourself: `go run ./cmd/rename -module github.com/you/myapp -name "My App" -bundle com.example.myapp`. `-identifier` overrides the Swift/Xcode name, which defaults to the display name in PascalCase (`MyApp`). The example notes feature (`internal/models/note.go`, `internal/db/queries/notes.go`, `pages/notes.templ`, the `/notes` routes in `internal/web/routes.go` and the nav link) is there to show the stack working end to end. Delete it when you no longer need it.

## Project layout

```text
cmd/server/          main package; only calls web.Run()
cmd/rename/          one-time rename tool for new projects (deletes itself)
internal/web/        server setup (server.go), routes, middleware, render helpers, JS import map
internal/db/         SQLite connection and migrations; queries/ (GORM CLI interfaces) and g/ (generated)
internal/models/     GORM models
components/, pages/  templ layout and pages
assets/              CSS, JS, vendored Hotwire libraries; velle-css lives in assets/css/velle and assets/js/velle
configurations/      Hotwire Native path configurations
ios/                 Hotwire Native iOS app
```

The server reads `assets/`, `configurations/` and `.importmap/` from the directory it runs in, so run it from the project root.

## Run locally

Requires Go 1.27.1 and a C compiler (for SQLite). templ, the GORM CLI and Air are pinned as Go tools in `go.mod`, so nothing else needs installing (the optional `velle` CLI is covered below). Run from the project root:

```sh
make dev              # live reload at http://localhost:7331
go run ./cmd/server   # or run once at http://localhost:3000
make gen              # regenerate templ and GORM CLI code
make test             # gen, vet and test
make build            # bin/server
```

`PORT` (default `3000`) and `DATABASE_PATH` (default `data/app.db`) configure the server. Turbo, Stimulus, and Hotwire Native Bridge are pinned in `internal/web/importmap.go`. The app serves their ES modules from `assets/vendor` through an inline import map, so JavaScript changes do not need a build step or Node.js. The versioned `.importmap` cache and public vendor files are included for offline startup. If public files are missing, `go-importmap` restores them from the cache; if the cache is missing, it downloads the pinned files from the CDN.

## Styles and components (velle-css)

Styling comes from [velle-css](https://github.com/codedbyshoe/velle-css) ([docs](https://codedbyshoe.github.io/velle-css/)): plain CSS with design tokens, a reset, base styles and utilities, plus copy-in components such as buttons, dialogs, tabs and form inputs. Its small JavaScript runtime drives interactive components through data attributes (`data-tabs`, `data-dialog`, …). It has no dependencies and picks up markup added at any time, so it works with Turbo navigation and Turbo Frames.

The starter includes only velle's core styles. Add components with the `velle` CLI, which reads `velle.json` and copies files into `assets/css/velle` and `assets/js/velle`:

```sh
go install github.com/codedbyshoe/velle-css/cmd/velle@latest
velle list                 # available components
velle add button dialog    # copy components (and their dependencies) into the project
```

You don't need to edit the layout. `assets/css/app.css` imports `assets/css/velle/main.css`, and `assets/js/app.js` imports `assets/js/velle/main.js`. `velle add` updates the import block between the `velle:begin` and `velle:end` markers in those two entry files and keeps it in load order. Don't edit inside that block; anything after `velle:end` is yours. Put site-specific styles in `assets/css/app.css` and use velle's tokens there, such as `var(--color-border)` and `var(--color-text-subtle)`. `velle add` never overwrites a file you've changed unless you pass `--overwrite`.

Some components are adapters for third-party libraries: `chart` (Chart.js), `datepicker` (flatpickr), `editor` (Tiptap), `inputmask` (maska) and `sortable` (SortableJS). velle copies the adapter but not the library, so add the library to `internal/web/importmap.go` the same way Turbo and Stimulus are pinned. Use the bare name the adapter imports (for example `sortablejs`) as the import map name. Some CDN ESM builds import further files of their own (Chart.js's jsDelivr `+esm` build imports `@kurkle/color`), so pin those too or use a self-contained build.

## Database

`internal/db` opens SQLite with GORM's official driver, [`gorm.io/driver/sqlite`](https://github.com/go-gorm/sqlite). It wraps `mattn/go-sqlite3`, so builds need CGO and a C compiler (Xcode's command line tools on macOS; the Dockerfile installs `gcc`). The connection string turns on WAL, `busy_timeout=5000`, foreign keys, `synchronous=NORMAL` and `_txlock=immediate`. `Open` in `internal/db/db.go` explains each setting.

The schema comes from `db.Migrate`, which runs `AutoMigrate` at startup. Add new models to that list. AutoMigrate only adds tables, columns and indexes; it never drops or renames anything. If a project needs versioned or destructive migrations, add a migration tool such as goose on top of the `*sql.DB` from `db.DB()`.

Typed queries use the [GORM CLI](https://github.com/go-gorm/cli). Write SQL-template interfaces in `internal/db/queries`, then run `make gen`. It generates the query methods and model field helpers into `internal/db/g`, and that code is committed:

```go
notes, err := g.Notes[models.Note](db).Recent(ctx, 50)
count, err := gorm.G[models.Note](db).Where(g.Note.Body.Eq("hi")).Count(ctx, "*")
```

## Deploy (Coolify)

The `Dockerfile` builds the server with CGO and an Alpine image that runs as a non-root user. In Coolify:

1. Create an application from the repository using the **Dockerfile** build pack and expose port `3000`.
2. Add persistent storage mounted at `/data`. The image defaults to `DATABASE_PATH=/data/app.db`.
3. Set the health check path to `/up`. It pings the database.
4. Optional backups with [Litestream](https://litestream.io): set `LITESTREAM_REPLICA_URL` (e.g. `s3://bucket/app.db`, adding `?endpoint=…&region=…` for non-AWS storage), `LITESTREAM_ACCESS_KEY_ID` and `LITESTREAM_SECRET_ACCESS_KEY`. With these set, the container restores the latest backup onto an empty volume and then runs the server under `litestream replicate`. Without them, the server runs on its own.

Run only one replica of the app. SQLite and Litestream both expect a single writer.

## Native app entry points

Point your Android or iOS Hotwire Native app's start location to the deployed site's `/` URL. The app serves versioned path configurations at:

- `/configurations/android_v1.json`
- `/configurations/ios_v1.json`

Bundle a copy of the corresponding JSON file in each mobile app and load the server URL as its remote configuration. This makes the navigation rules available offline and lets you update them later. Android's default rule maps web screens to `hotwire://fragment/web`, so register `HotwireWebFragment` as a fragment destination. For example, in an Android `Application`:

```kotlin
Hotwire.registerFragmentDestinations(HotwireWebFragment::class)
Hotwire.loadPathConfiguration(
    context = this,
    location = PathConfiguration.Location(
        assetFilePath = "json/configuration.json",
        remoteFileUrl = "https://your-domain.example/configurations/android_v1.json"
    )
)
```

The included iOS app lives in `ios/CodedByShoe.xcodeproj`. It uses the same `configurations/ios_v1.json` file as its bundled path configuration and loads the server copy too. Its Swift `button` bridge component displays the home page's “About” link in the native navigation bar. Android still needs its own wrapper and native bridge component; see the [Android bridge guide](https://native.hotwired.dev/android/bridge-components).

Pages requested with a Hotwire Native Android or iOS user agent omit the web header, leaving navigation chrome to the mobile app.

## Test on an iPhone simulator

Start the server from the project root with the commands above. Open `ios/CodedByShoe.xcodeproj` in Xcode, choose an iPhone simulator, and run the `CodedByShoe` scheme. The app currently uses `http://localhost:3000/`, which reaches the Mac from the iOS Simulator. The `Info.plist` exception allows insecure HTTP only for localhost during this test.

You can also run the navigation UI test from Terminal:

```sh
xcodebuild -project ios/CodedByShoe.xcodeproj \
  -scheme CodedByShoe \
  -destination 'platform=iOS Simulator,name=iPhone 17' \
  CODE_SIGNING_ALLOWED=NO test
```

For a physical iPhone, change the base URL in `ios/CodedByShoe/AppEnvironment.swift` to a reachable HTTPS host. An Android emulator can reach a server running on your computer at `http://10.0.2.2:3000`; physical Android devices also need a reachable host or tunnel.

## Check

```sh
make test
go test ./internal/web -run TestHotwireRoutes/about   # a single test
```

`internal/web/web_test.go` tests the routes through `httptest` against a temporary database, and checks the rendered HTML for page titles, import map entries and Hotwire attributes. Update its expected strings when you change page copy or vendored library paths. `internal/db/db_test.go` covers the SQLite settings, the generated queries and concurrent writes.

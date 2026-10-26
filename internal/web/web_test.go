package web

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"

	"github.com/codedbyshoe/go-hotwire-starter/internal/db"
)

func testApp(t *testing.T) http.Handler {
	t.Helper()
	// The server reads assets/ and configurations/ from the working
	// directory, and tests run from this package's folder, so move to the
	// repo root first.
	t.Chdir("../..")

	importMap, err := hotwireImportMap(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	database, err := db.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Migrate(database); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close(database) })
	return newServer(importMap, database)
}

func serve(app http.Handler, request *http.Request) *http.Response {
	recorder := httptest.NewRecorder()
	app.ServeHTTP(recorder, request)
	return recorder.Result()
}

func TestHotwireRoutes(t *testing.T) {
	app := testApp(t)

	for _, tc := range []struct {
		path     string
		contains []string
	}{
		{"/", []string{"<title>Home · Coded by Shoe</title>", "id=\"server-time\"", "src=\"/time\"", "data-controller=\"bridge--button\"", `type="importmap"`, `"@hotwired/turbo": "/assets/vendor/turbo.js/turbo.js"`, `"@hotwired/stimulus": "/assets/vendor/stimulus.js/stimulus.js"`, `"@hotwired/hotwire-native-bridge": "/assets/vendor/hotwire-native-bridge.js/hotwire-native-bridge.js"`, "/assets/js/app.js"}},
		{"/about", []string{"<title>About · Coded by Shoe</title>", "One set of HTML routes."}},
		{"/notes", []string{"<title>Notes · Coded by Shoe</title>", "No notes yet."}},
		{"/up", []string{"ok"}},
		{"/time", []string{"<turbo-frame id=\"server-time\"", "data-action=\"refresh#refresh\""}},
		{"/assets/js/app.js", []string{"@hotwired/turbo", "@hotwired/stimulus"}},
		{"/assets/vendor/turbo.js/turbo.js", []string{"Turbo"}},
		{"/assets/vendor/stimulus.js/stimulus.js", []string{"Controller"}},
		{"/assets/vendor/hotwire-native-bridge.js/hotwire-native-bridge.js", []string{"BridgeComponent"}},
	} {
		t.Run(tc.path, func(t *testing.T) {
			response := serve(app, httptest.NewRequest("GET", tc.path, nil))
			defer response.Body.Close()
			if response.StatusCode != 200 {
				t.Fatalf("status = %d", response.StatusCode)
			}
			body, err := io.ReadAll(response.Body)
			if err != nil {
				t.Fatal(err)
			}
			for _, want := range tc.contains {
				if !strings.Contains(string(body), want) {
					t.Errorf("body does not contain %q", want)
				}
			}
		})
	}
}

func TestNativePathConfigurations(t *testing.T) {
	app := testApp(t)
	for _, platform := range []string{"android", "ios"} {
		t.Run(platform, func(t *testing.T) {
			response := serve(app, httptest.NewRequest("GET", "/configurations/"+platform+"_v1.json", nil))
			defer response.Body.Close()
			if response.StatusCode != 200 {
				t.Fatalf("status = %d", response.StatusCode)
			}
			var config struct {
				Settings map[string]any `json:"settings"`
				Rules    []struct {
					Patterns   []string       `json:"patterns"`
					Properties map[string]any `json:"properties"`
				} `json:"rules"`
			}
			if err := json.NewDecoder(response.Body).Decode(&config); err != nil {
				t.Fatal(err)
			}
			if len(config.Rules) == 0 || len(config.Rules[0].Patterns) == 0 || config.Rules[0].Properties["context"] != "default" {
				t.Fatal("missing default navigation rule")
			}
			if platform == "android" && config.Rules[0].Properties["uri"] != "hotwire://fragment/web" {
				t.Fatal("Android web destination URI is missing")
			}
		})
	}
}

func TestNativeWebViewGetsNativeLayout(t *testing.T) {
	for _, platform := range []string{"Android", "iOS"} {
		t.Run(platform, func(t *testing.T) {
			request := httptest.NewRequest("GET", "/", nil)
			request.Header.Set("User-Agent", "Example; Hotwire Native "+platform+"; Turbo Native "+platform+";")
			response := serve(testApp(t), request)
			defer response.Body.Close()
			body, err := io.ReadAll(response.Body)
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(string(body), `class="hotwire-native"`) {
				t.Fatal("native layout class is missing")
			}
		})
	}
}

func TestCreateNote(t *testing.T) {
	app := testApp(t)

	form := url.Values{"body": {"Hello from the test"}}
	request := httptest.NewRequest("POST", "/notes", strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response := serve(app, request)
	if response.StatusCode != http.StatusSeeOther || response.Header.Get("Location") != "/notes" {
		t.Fatalf("status = %d, location = %q", response.StatusCode, response.Header.Get("Location"))
	}

	body, _ := io.ReadAll(serve(app, httptest.NewRequest("GET", "/notes", nil)).Body)
	if !strings.Contains(string(body), "Hello from the test") {
		t.Fatal("created note is not listed")
	}
}

func TestCreateEmptyNoteIsRejected(t *testing.T) {
	request := httptest.NewRequest("POST", "/notes", strings.NewReader("body=+"))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response := serve(testApp(t), request)
	if response.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d", response.StatusCode)
	}
}

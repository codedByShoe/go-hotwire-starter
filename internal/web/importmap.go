package web

import (
	"context"
	"fmt"

	importmap "github.com/donseba/go-importmap"
	"github.com/donseba/go-importmap/client/raw"
	"github.com/donseba/go-importmap/library"
)

// Ship the cached files and their public copies with the app so startup does
// not depend on a CDN. CacheOrFetch restores missing public files from cache.
func hotwireImportMap(ctx context.Context) (string, error) {
	packages := []library.Package{
		{
			Name: "turbo.js", Version: "8.0.23",
			Provider: raw.New("https://cdn.jsdelivr.net/npm/@hotwired/turbo@8.0.23/dist/turbo.es2017-esm.js"),
			Require:  library.Includes{{File: "turbo.js", As: "@hotwired/turbo"}},
		},
		{
			Name: "stimulus.js", Version: "3.2.2",
			Provider: raw.New("https://cdn.jsdelivr.net/npm/@hotwired/stimulus@3.2.2/dist/stimulus.js"),
			Require:  library.Includes{{File: "stimulus.js", As: "@hotwired/stimulus"}},
		},
		{
			Name: "hotwire-native-bridge.js", Version: "1.2.2",
			Provider: raw.New("https://cdn.jsdelivr.net/npm/@hotwired/hotwire-native-bridge@1.2.2/dist/hotwire-native-bridge.js"),
			Require:  library.Includes{{File: "hotwire-native-bridge.js", As: "@hotwired/hotwire-native-bridge"}},
		},
	}

	im := importmap.NewDefaults().
		AssetsDir("assets/vendor").
		ShimPath("").
		WithPackages(packages)
	if err := im.CacheOrFetch(ctx); err != nil {
		return "", fmt.Errorf("prepare Hotwire import map: %w", err)
	}
	head, err := im.Render()
	if err != nil {
		return "", fmt.Errorf("render Hotwire import map: %w", err)
	}
	return string(head), nil
}

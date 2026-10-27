// Command rename turns a fresh copy of the starter into your project: it
// replaces the starter's module path, display name, Swift identifier and iOS
// bundle ID everywhere, renames the matching files and directories,
// regenerates templ code, and then deletes itself. Run it once from the
// project root:
//
//	go run ./cmd/rename -name "My App" -bundle com.example.myapp
//
// After gonew, the new module path is read from go.mod. gonew only rewrites
// imports in .go files, so this also fixes the imports in .templ files. On a
// plain clone, pass -module github.com/you/myapp to set it.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"unicode"
)

// The placeholders the starter ships with.
const (
	oldModule     = "github.com/codedbyshoe/go-hotwire-starter"
	oldName       = "Coded by Shoe"
	oldIdentifier = "CodedByShoe"
	oldBundle     = "com.codedbyshoe.hotwire"
)

// Directories never rewritten: vendored JS, build output, local data, and
// per-user Xcode state.
var skipDirs = []string{".git", ".importmap", "vendor", "tmp", "bin", "data", "node_modules", "xcuserdata", "DerivedData"}

func main() {
	name := flag.String("name", "", `display name, e.g. "My App" (required)`)
	bundle := flag.String("bundle", "", "iOS bundle identifier, e.g. com.example.myapp (required)")
	identifier := flag.String("identifier", "", "Swift/Xcode identifier (default: name in PascalCase)")
	module := flag.String("module", "", "Go module path, e.g. github.com/you/myapp (default: the module in go.mod)")
	flag.Parse()

	if *name == "" || *bundle == "" {
		flag.Usage()
		os.Exit(2)
	}
	if *identifier == "" {
		*identifier = pascalCase(*name)
	}
	if !regexp.MustCompile(`^[A-Za-z][A-Za-z0-9]*$`).MatchString(*identifier) {
		fail(fmt.Errorf("identifier %q must be letters and digits, starting with a letter; pass -identifier", *identifier))
	}
	if !regexp.MustCompile(`^[A-Za-z0-9-]+(\.[A-Za-z0-9-]+)+$`).MatchString(*bundle) {
		fail(fmt.Errorf("bundle %q is not a reverse-DNS identifier", *bundle))
	}
	if *module == "" {
		path, err := readModulePath("go.mod")
		if err != nil {
			fail(fmt.Errorf("run from the project root: %w", err))
		}
		*module = path
	}
	if strings.ContainsAny(*module, " \t") {
		fail(fmt.Errorf("module %q must not contain spaces", *module))
	}

	nameReplacer := strings.NewReplacer(
		oldBundle, *bundle,
		oldName, *name,
		oldIdentifier, *identifier,
	)
	// Markdown keeps the starter's module path, so the README's gonew
	// instructions still point at the starter.
	moduleReplacer := strings.NewReplacer(oldModule, *module)

	self := filepath.Join("cmd", "rename")
	var renames []string
	err := filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && (slices.Contains(skipDirs, d.Name()) || path == self) {
			return filepath.SkipDir
		}
		if strings.Contains(d.Name(), oldIdentifier) {
			renames = append(renames, path)
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Ext(path) == ".md" {
			return rewrite(path, nameReplacer)
		}
		return rewrite(path, nameReplacer, moduleReplacer)
	})
	if err != nil {
		fail(err)
	}

	// Rename deepest paths first so parent renames don't invalidate them.
	slices.Reverse(renames)
	for _, path := range renames {
		dir, base := filepath.Split(path)
		target := filepath.Join(dir, strings.ReplaceAll(base, oldIdentifier, *identifier))
		if err := os.Rename(path, target); err != nil {
			fail(err)
		}
		fmt.Println("renamed", path, "→", target)
	}

	cmd := exec.Command("go", "tool", "templ", "generate")
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		fail(fmt.Errorf("templ generate: %w", err))
	}

	if err := os.RemoveAll(self); err != nil {
		fail(err)
	}
	fmt.Printf("Renamed to %q (%s, %s, %s). Next: go test ./...\n", *name, *module, *identifier, *bundle)
}

func rewrite(path string, replacers ...*strings.Replacer) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	// Skip binaries such as images and Xcode's user state.
	if bytes.IndexByte(data, 0) >= 0 {
		return nil
	}
	updated := string(data)
	for _, replacer := range replacers {
		updated = replacer.Replace(updated)
	}
	if updated == string(data) {
		return nil
	}
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	fmt.Println("updated", path)
	return os.WriteFile(path, []byte(updated), info.Mode().Perm())
}

// readModulePath returns the module path from the "module" line of go.mod.
func readModulePath(goModPath string) (string, error) {
	data, err := os.ReadFile(goModPath)
	if err != nil {
		return "", err
	}
	for _, line := range strings.Split(string(data), "\n") {
		if path, ok := strings.CutPrefix(strings.TrimSpace(line), "module "); ok {
			return strings.Trim(strings.TrimSpace(path), `"`), nil
		}
	}
	return "", fmt.Errorf("no module line in %s", goModPath)
}

func pascalCase(s string) string {
	var b strings.Builder
	upper := true
	for _, r := range s {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			if upper {
				r = unicode.ToUpper(r)
			}
			b.WriteRune(r)
			upper = false
		default:
			upper = true
		}
	}
	return b.String()
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "rename:", err)
	os.Exit(1)
}

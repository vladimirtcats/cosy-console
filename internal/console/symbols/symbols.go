// Package symbols exposes cosy-console packages to the console session.
//
// Only types the session user constructs are registered: domain params and
// models. Repositories, usecases and infrastructure arrive as live values
// through the App export and need no symbols.
//
// # Regeneration
//
// After changing the exported surface of a domain, re-run extract per
// package and merge the outputs into the per-domain file:
//
//	go run github.com/traefik/yaegi/cmd/yaegi extract -name symbols cosy-console/internal/domain/orders
//	go run github.com/traefik/yaegi/cmd/yaegi extract -name symbols cosy-console/internal/domain/orders/models
//
//	# merge both files into orders.go and keep the raw extract keys as is —
//	# re-keying for the session happens in Exports(), in code
//
// extract writes one file per package into the working directory; delete
// the raw files after merging.
package symbols

import (
	"fmt"
	"path"
	"strings"

	"github.com/traefik/yaegi/interp"
)

// Symbols holds the binary symbols of cosy-console packages for the Yaegi
// interpreter, each under the raw key produced by yaegi extract:
// path.Join(importPath, packageName). It is populated by generated init()
// functions in this package; do not edit the keys by hand.
var Symbols = interp.Exports{}

// Exports returns the symbol table re-keyed for the console session.
//
// ImportUsed binds a registered package by the LAST segment of its key, so
// a key must end with the package name and the name must be unique. Raw
// extract keys fail both ways: sibling models packages all end in "models"
// and collide, and a domain path may end in a segment that differs from
// its package name. Exports therefore re-keys every package to
// console/<name>/<name>, where the name is derived from the raw key:
//
//   - a top-level package (a domain, pagination, uuid) is bound under its
//     own package name;
//   - a package nested inside another registered package (the models
//     package of a domain) is bound under <parent>_<name>, e.g.
//     orders_models — unique, and friendlier at the prompt than colliding
//     packages named "models".
//
// Re-keying happens here, in code, so the generated files keep their raw
// extract keys and are never edited by hand.
func Exports() (interp.Exports, error) {
	paths := importPaths()

	exports := make(interp.Exports, len(Symbols))
	owners := make(map[string]string, len(Symbols))

	for key, syms := range Symbols {
		name := path.Base(key)

		if parent, ok := parentPackage(importPath(key), paths); ok {
			name = path.Base(parent) + "_" + name
		}

		if prev, dup := owners[name]; dup {
			return nil, fmt.Errorf("console session name %q is claimed by both %s and %s", name, prev, key)
		}

		owners[name] = key
		exports["console/"+name+"/"+name] = syms
	}

	return exports, nil
}

// importPath returns the package path behind a raw extract key: extract
// registers a package as path.Join(importPath, packageName).
func importPath(key string) string {
	return strings.TrimSuffix(key, "/"+path.Base(key))
}

func importPaths() []string {
	paths := make([]string, 0, len(Symbols))

	for key := range Symbols {
		paths = append(paths, importPath(key))
	}

	return paths
}

// parentPackage reports the closest registered package that contains
// pkgPath as a subdirectory, e.g. the orders domain for orders/models.
func parentPackage(pkgPath string, paths []string) (string, bool) {
	parent := ""

	for _, candidate := range paths {
		if candidate == pkgPath || !strings.HasPrefix(pkgPath, candidate+"/") {
			continue
		}

		if len(candidate) > len(parent) {
			parent = candidate
		}
	}

	return parent, parent != ""
}

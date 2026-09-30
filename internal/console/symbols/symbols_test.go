package symbols_test

import (
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/traefik/yaegi/interp"

	"cosy-console/internal/console/symbols"
)

func TestExports_BindsEveryPackageUnderAUniqueConsoleName(t *testing.T) {
	tests := []struct {
		name string
		key  string
	}{
		{name: "orders domain is bound under its package name", key: "console/orders/orders"},
		{name: "orders models is bound under the parent-prefixed name", key: "console/orders_models/orders_models"},
		{name: "catalog domain", key: "console/catalog/catalog"},
		{name: "pagination", key: "console/pagination/pagination"},
		{name: "uuid", key: "console/uuid/uuid"},
	}

	exports, err := symbols.Exports()

	require.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			syms, ok := exports[tt.key]
			require.True(t, ok, "expected %q in the export table, got keys %v", tt.key, keysOf(exports))
			assert.NotEmpty(t, syms)
		})
	}
}

func TestExports_EveryKeyFollowsTheConsoleNameShape(t *testing.T) {
	exports, err := symbols.Exports()

	require.NoError(t, err)
	require.NotEmpty(t, exports)

	for key := range exports {
		name := path.Base(key)

		assert.Equal(t, "console/"+name+"/"+name, key,
			"every key must follow the console/<name>/<name> shape")
	}
}

func TestExports_FailsFastOnDuplicateSessionNames(t *testing.T) {
	original := symbols.Symbols
	t.Cleanup(func() { symbols.Symbols = original })

	// Two unrelated models packages: neither is nested in a registered
	// parent, so both claim the bare name and must be rejected.
	symbols.Symbols = interp.Exports{
		"myapp/internal/orders/models/models":  {},
		"myapp/internal/catalog/models/models": {},
	}

	_, err := symbols.Exports()

	require.Error(t, err)
	assert.Contains(t, err.Error(), `"models"`)
}

func keysOf(exports interp.Exports) []string {
	keys := make([]string, 0, len(exports))

	for key := range exports {
		keys = append(keys, key)
	}

	return keys
}

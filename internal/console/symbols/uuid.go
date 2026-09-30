package symbols

import (
	"reflect"

	"github.com/google/uuid"
)

// The uuid symbol set is hand-maintained and deliberately minimal: only the
// API the application itself uses.
func init() {
	Symbols["github.com/google/uuid/uuid"] = map[string]reflect.Value{
		"Must":      reflect.ValueOf(uuid.Must),
		"MustParse": reflect.ValueOf(uuid.MustParse),
		"New":       reflect.ValueOf(uuid.New),
		"NewV7":     reflect.ValueOf(uuid.NewV7),
		"Parse":     reflect.ValueOf(uuid.Parse),

		"Nil":  reflect.ValueOf(&uuid.Nil).Elem(),
		"UUID": reflect.ValueOf((*uuid.UUID)(nil)),
	}
}

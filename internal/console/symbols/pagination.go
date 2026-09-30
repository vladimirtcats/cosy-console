package symbols

import (
	"reflect"

	"cosy-console/pkg/pagination"
)

func init() {
	Symbols["cosy-console/pkg/pagination/pagination"] = map[string]reflect.Value{
		"NewSettings":  reflect.ValueOf(pagination.NewSettings),
		"WithPage":     reflect.ValueOf(pagination.WithPage),
		"WithPageSize": reflect.ValueOf(pagination.WithPageSize),

		"MaxPageSize": reflect.ValueOf(pagination.MaxPageSize),
		"Settings":    reflect.ValueOf((*pagination.Settings)(nil)),
		"Result":      reflect.ValueOf((*pagination.Result)(nil)),
	}
}

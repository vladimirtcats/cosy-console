package catalog

type ItemFilter struct {
	SKU             string
	IncludeInactive bool
}

func (f ItemFilter) IsDefault() bool {
	return f.SKU == "" && !f.IncludeInactive
}

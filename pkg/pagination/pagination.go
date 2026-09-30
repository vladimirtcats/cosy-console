package pagination

const MaxPageSize = 100

type Settings struct {
	Page     int
	PageSize int
}

type Option func(*Settings)

func WithPage(page int) Option {
	return func(s *Settings) {
		s.Page = page
	}
}

func WithPageSize(pageSize int) Option {
	return func(s *Settings) {
		s.PageSize = pageSize
	}
}

func NewSettings(opts ...Option) Settings {
	s := Settings{}

	for _, opt := range opts {
		opt(&s)
	}

	s.Normalize()

	return s
}

func (s *Settings) Normalize() {
	if s.Page < 1 {
		s.Page = 1
	}

	if s.PageSize <= 0 {
		s.PageSize = 20
	}

	if s.PageSize > MaxPageSize {
		s.PageSize = MaxPageSize
	}
}

func (s Settings) Offset() int {
	return (s.Page - 1) * s.PageSize
}

type Result struct {
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
}

func NewResult(total int64, s Settings) Result {
	return Result{
		Total:      total,
		TotalPages: totalPages(total, s.PageSize),
		Page:       s.Page,
		PageSize:   s.PageSize,
	}
}

func totalPages(total int64, pageSize int) int {
	if pageSize < 1 {
		return 0
	}

	return int((total + int64(pageSize) - 1) / int64(pageSize))
}

package pagination

// Params represents pagination and sorting parameters
type Params struct {
	Limit         int    // Number of items per page
	Offset        int    // Starting position
	SortBy        string // Field to sort by
	SortDirection string // Sort direction (asc or desc)
}

// NewParams creates a new pagination parameters object with defaults
func NewParams(limit, offset int, sortBy, sortDirection string) Params {
	if limit <= 0 {
		limit = 10 // Default limit
	}

	if offset < 0 {
		offset = 0 // Default offset
	}

	// Default sort direction
	if sortDirection != "asc" && sortDirection != "desc" {
		sortDirection = "desc"
	}

	return Params{
		Limit:         limit,
		Offset:        offset,
		SortBy:        sortBy,
		SortDirection: sortDirection,
	}
}

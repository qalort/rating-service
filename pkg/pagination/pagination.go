package pagination

// Params interface defines methods for pagination parameters
type Params interface {
        GetPage() int
        GetLimit() int
        GetOffset() int
        GetSortBy() string
        GetSortDirection() string
}

// PaginationParams represents pagination and sorting parameters
type PaginationParams struct {
        Page          int    // Page number (1-based)
        Limit         int    // Number of items per page
        SortBy        string // Field to sort by
        SortDirection string // Sort direction (asc or desc)
}

// GetPage returns the page number
func (p PaginationParams) GetPage() int {
        return p.Page
}

// GetLimit returns the number of items per page
func (p PaginationParams) GetLimit() int {
        return p.Limit
}

// GetOffset returns the offset based on page and limit
func (p PaginationParams) GetOffset() int {
        return (p.Page - 1) * p.Limit
}

// GetSortBy returns the field to sort by
func (p PaginationParams) GetSortBy() string {
        return p.SortBy
}

// GetSortDirection returns the sort direction
func (p PaginationParams) GetSortDirection() string {
        return p.SortDirection
}

// NewParams creates a new pagination parameters object with defaults
func NewParams(page, limit int, sortBy, sortDirection string) Params {
        if page <= 0 {
                page = 1 // Default page
        }

        if limit <= 0 {
                limit = 10 // Default limit
        }

        // Default sort direction
        if sortDirection != "asc" && sortDirection != "desc" {
                sortDirection = "desc"
        }

        return &PaginationParams{
                Page:          page,
                Limit:         limit,
                SortBy:        sortBy,
                SortDirection: sortDirection,
        }
}

// NewParamsWithOffset creates a new pagination parameters object with direct offset value
func NewParamsWithOffset(limit, offset int, sortBy, sortDirection string) Params {
        if limit <= 0 {
                limit = 10 // Default limit
        }

        if offset < 0 {
                offset = 0 // Default offset
        }

        // Calculate page from offset and limit
        page := 1
        if offset > 0 && limit > 0 {
                page = (offset / limit) + 1
        }

        // Default sort direction
        if sortDirection != "asc" && sortDirection != "desc" {
                sortDirection = "desc"
        }

        return &PaginationParams{
                Page:          page,
                Limit:         limit,
                SortBy:        sortBy,
                SortDirection: sortDirection,
        }
}

// Pagination is a concrete implementation of pagination
type Pagination struct {
        Page  int
        Limit int
}

// NewPagination creates a new Pagination instance
func NewPagination(page, limit int) *Pagination {
        if page <= 0 {
                page = 1
        }
        if limit <= 0 {
                limit = 10
        }
        return &Pagination{
                Page:  page,
                Limit: limit,
        }
}

// GetLimit returns the number of items per page
func (p *Pagination) GetLimit() int {
        return p.Limit
}

// GetOffset calculates offset based on page and limit
func (p *Pagination) GetOffset() int {
        return (p.Page - 1) * p.Limit
}

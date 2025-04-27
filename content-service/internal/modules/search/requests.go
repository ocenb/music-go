package search

type SearchForm struct {
	Query string `form:"query" binding:"required"`
}

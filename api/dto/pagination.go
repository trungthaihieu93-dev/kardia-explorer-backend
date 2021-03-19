// Package dto
package dto

type Pagination struct {
	Page  int64 `json:"page"`
	Limit int64 `json:"limit"`
	Total int64 `json:"total"`
}

func (p Pagination) ToMongoFilter() (int64, int64) {
	skip := (p.Page - 1) * p.Limit
	limit := p.Limit
	return skip, limit

}

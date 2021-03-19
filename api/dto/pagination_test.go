// Package dto
package dto

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPagination_MongoFilter(t *testing.T) {
	skip1, _ := Pagination{Page: 1, Limit: 10}.ToMongoFilter()
	assert.Equal(t, 0, skip1)
	skip2, _ := Pagination{Page: 2, Limit: 10}.ToMongoFilter()
	assert.Equal(t, 10, skip2)
	skip3, _ := Pagination{Page: 3, Limit: 10}.ToMongoFilter()
	assert.Equal(t, 20, skip3)
	skip4, _ := Pagination{Page: 4, Limit: 10}.ToMongoFilter()
	assert.Equal(t, 30, skip4)
}

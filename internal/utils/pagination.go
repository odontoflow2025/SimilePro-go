package utils

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PaginationMeta struct {
	CurrentPage  int   `json:"current_page"`
	Limit        int   `json:"limit"`
	TotalRecords int64 `json:"total_records"`
}

type PaginatedResponse struct {
	Data interface{}    `json:"data"`
	Meta PaginationMeta `json:"meta"`
}

// GetPaginationParams extrai page e limit da querystring, com fallbacks seguros.
func GetPaginationParams(c *gin.Context) (int, int) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page <= 0 {
		page = 1
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	switch {
	case limit <= 0:
		limit = 20
	case limit > 50:
		limit = 50
	}

	return page, limit
}

// Paginate retorna um scope para o GORM aplicar OFFSET e LIMIT.
func Paginate(page, limit int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		offset := (page - 1) * limit
		return db.Offset(offset).Limit(limit)
	}
}

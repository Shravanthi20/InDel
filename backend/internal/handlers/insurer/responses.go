package insurer

import (
	"time"

	"github.com/gin-gonic/gin"
)

type Meta struct {
	RequestID string `json:"request_id"`
	Timestamp string `json:"timestamp"`
}

type SuccessResponse struct {
	Data interface{} `json:"data"`
	Meta Meta        `json:"meta"`
}

type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Pagination Pagination  `json:"pagination"`
	Meta       Meta        `json:"meta"`
}

type Pagination struct {
	Page    int  `json:"page"`
	Limit   int  `json:"limit"`
	Total   int  `json:"total"`
	HasNext bool `json:"has_next"`
}

type ErrorDetail struct {
	Field string `json:"field"`
}

type ErrorInfo struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details ErrorDetail `json:"details,omitempty"`
}

type ErrorResponse struct {
	Error ErrorInfo `json:"error"`
	Meta  Meta      `json:"meta"`
}

func getMeta(c *gin.Context) Meta {
	reqID := c.GetHeader("X-Request-ID")
	if reqID == "" {
		reqID = "req_123" // Fallback or generate
	}
	return Meta{
		RequestID: reqID,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
}

func SendSuccess(c *gin.Context, data interface{}) {
	c.JSON(200, SuccessResponse{
		Data: data,
		Meta: getMeta(c),
	})
}

func SendPaginated(c *gin.Context, data interface{}, page, limit, total int) {
	hasNext := (page * limit) < total
	c.JSON(200, PaginatedResponse{
		Data: data,
		Pagination: Pagination{
			Page:    page,
			Limit:   limit,
			Total:   total,
			HasNext: hasNext,
		},
		Meta: getMeta(c),
	})
}

func SendError(c *gin.Context, status int, code, message, field string) {
	c.JSON(status, ErrorResponse{
		Error: ErrorInfo{
			Code:    code,
			Message: message,
			Details: ErrorDetail{Field: field},
		},
		Meta: getMeta(c),
	})
}

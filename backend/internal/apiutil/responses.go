package apiutil

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

type Pagination struct {
	Page    int  `json:"page"`
	Limit   int  `json:"limit"`
	Total   int  `json:"total"`
	HasNext bool `json:"has_next"`
}

type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Pagination Pagination  `json:"pagination"`
	Meta       Meta        `json:"meta"`
}

type ErrorDetail struct {
	Field string `json:"field,omitempty"`
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

func meta(c *gin.Context) Meta {
	requestID := c.GetHeader("X-Request-ID")
	if requestID == "" {
		requestID = "req_123"
	}

	return Meta{
		RequestID: requestID,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
}

func SendSuccess(c *gin.Context, status int, data interface{}) {
	c.JSON(status, SuccessResponse{
		Data: data,
		Meta: meta(c),
	})
}

func SendPaginated(c *gin.Context, data interface{}, page, limit, total int) {
	c.JSON(200, PaginatedResponse{
		Data: data,
		Pagination: Pagination{
			Page:    page,
			Limit:   limit,
			Total:   total,
			HasNext: page*limit < total,
		},
		Meta: meta(c),
	})
}

func SendError(c *gin.Context, status int, code, message, field string) {
	c.JSON(status, ErrorResponse{
		Error: ErrorInfo{
			Code:    code,
			Message: message,
			Details: ErrorDetail{Field: field},
		},
		Meta: meta(c),
	})
}

package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response is the standard API response envelope
type Response struct {
	Success   bool        `json:"success"`
	Data      interface{} `json:"data,omitempty"`
	Error     *ErrorInfo  `json:"error,omitempty"`
	Meta      *Meta       `json:"meta,omitempty"`
	RequestID string      `json:"request_id,omitempty"`
}

// ErrorInfo holds error details
type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// Meta holds pagination metadata
type Meta struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
	HasNext    bool  `json:"has_next"`
	HasPrev    bool  `json:"has_prev"`
}

// OK sends a 200 success response
func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Success:   true,
		Data:      data,
		RequestID: c.GetString("request_id"),
	})
}

// Created sends a 201 created response
func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, Response{
		Success:   true,
		Data:      data,
		RequestID: c.GetString("request_id"),
	})
}

// Paginated sends a paginated response
func Paginated(c *gin.Context, data interface{}, meta Meta) {
	c.JSON(http.StatusOK, Response{
		Success:   true,
		Data:      data,
		Meta:      &meta,
		RequestID: c.GetString("request_id"),
	})
}

// Fail sends an error response
func Fail(c *gin.Context, status int, code, message string) {
	c.JSON(status, Response{
		Success:   false,
		Error:     &ErrorInfo{Code: code, Message: message},
		RequestID: c.GetString("request_id"),
	})
}

// FailDetail sends an error response with additional detail
func FailDetail(c *gin.Context, status int, code, message, details string) {
	c.JSON(status, Response{
		Success: false,
		Error:   &ErrorInfo{Code: code, Message: message, Details: details},
	})
}

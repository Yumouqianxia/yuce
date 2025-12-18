// Package response provides standardized HTTP response utilities for REST APIs.
//
// This package implements a consistent response format for all API endpoints,
// ensuring uniform structure for both successful and error responses. It integrates
// seamlessly with the Gin web framework and provides convenient helper functions
// for common HTTP status codes and response patterns.
//
// Key Features:
//   - Standardized response structure for consistency
//   - Type-safe error handling with detailed error information
//   - Convenient helper functions for common HTTP status codes
//   - Integration with Gin framework for easy usage
//   - Support for both data and error responses
//   - Structured error details for better debugging
//
// Response Format:
// All responses follow a consistent JSON structure:
//
//	{
//	  "success": true|false,
//	  "message": "Human readable message",
//	  "data": {...},           // Present on success
//	  "error": {               // Present on error
//	    "code": 400,
//	    "message": "Error message",
//	    "details": "Additional details"
//	  }
//	}
//
// Example usage:
//
//	// Success response
//	response.OK(c, "User retrieved successfully", user)
//
//	// Error response
//	response.BadRequest(c, "Invalid user ID format")
//
//	// Custom error with details
//	response.ValidationError(c, "Email field is required")
package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response represents the standardized API response structure used across all endpoints.
//
// This structure ensures consistency in API responses and provides clear indication
// of operation success or failure. The response includes a success flag, human-readable
// message, optional data payload, and detailed error information when applicable.
//
// Fields:
//   - Success: Boolean flag indicating if the operation was successful
//   - Message: Human-readable message describing the operation result
//   - Data: Optional payload containing the response data (only present on success)
//   - Error: Optional error information (only present on failure)
//
// JSON Tags:
// The struct uses JSON tags with examples for automatic API documentation generation.
type Response struct {
	Success bool        `json:"success" example:"true"`                 // Operation success status
	Message string      `json:"message" example:"Operation successful"` // Human-readable message
	Data    interface{} `json:"data,omitempty"`                         // Response payload (success only)
	Error   *ErrorInfo  `json:"error,omitempty"`                        // Error details (failure only)
}

// ErrorInfo represents detailed error information for failed API operations.
//
// This structure provides comprehensive error details including HTTP status code,
// error message, and additional context for debugging and client-side error handling.
//
// Fields:
//   - Code: HTTP status code corresponding to the error
//   - Message: Primary error message describing what went wrong
//   - Details: Additional context or validation details about the error
//
// The error information is designed to be both machine-readable (via codes)
// and human-readable (via messages and details).
type ErrorInfo struct {
	Code    int    `json:"code" example:"400"`                            // HTTP status code
	Message string `json:"message" example:"Bad request"`                 // Primary error message
	Details string `json:"details,omitempty" example:"Validation failed"` // Additional error context
}

// Success sends a successful HTTP response with the specified status code, message, and data.
//
// This function creates a standardized success response and sends it as JSON.
// It sets the Success field to true and includes the provided data payload.
//
// Parameters:
//   - c: Gin context for the HTTP request
//   - statusCode: HTTP status code (typically 200, 201, etc.)
//   - message: Human-readable success message
//   - data: Response payload (can be nil for responses without data)
//
// Example:
//
//	user := &User{ID: 1, Name: "John"}
//	response.Success(c, http.StatusOK, "User retrieved successfully", user)
//
// This will produce a JSON response like:
//
//	{
//	  "success": true,
//	  "message": "User retrieved successfully",
//	  "data": {"id": 1, "name": "John"}
//	}
func Success(c *gin.Context, statusCode int, message string, data interface{}) {
	c.JSON(statusCode, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// Error sends an error HTTP response with the specified status code, message, and details.
//
// This function creates a standardized error response and sends it as JSON.
// It sets the Success field to false and includes detailed error information.
//
// Parameters:
//   - c: Gin context for the HTTP request
//   - statusCode: HTTP status code (400, 404, 500, etc.)
//   - message: Primary error message
//   - details: Additional error context or validation details
//
// Example:
//
//	response.Error(c, http.StatusBadRequest, "Invalid input", "Email format is invalid")
//
// This will produce a JSON response like:
//
//	{
//	  "success": false,
//	  "message": "Invalid input",
//	  "error": {
//	    "code": 400,
//	    "message": "Invalid input",
//	    "details": "Email format is invalid"
//	  }
//	}
func Error(c *gin.Context, statusCode int, message string, details string) {
	c.JSON(statusCode, Response{
		Success: false,
		Message: message,
		Error: &ErrorInfo{
			Code:    statusCode,
			Message: message,
			Details: details,
		},
	})
}

// ValidationError 验证错误响应
func ValidationError(c *gin.Context, details string) {
	Error(c, http.StatusBadRequest, "Validation failed", details)
}

// NotFound 资源不存在响应
func NotFound(c *gin.Context, resource string) {
	Error(c, http.StatusNotFound, "Resource not found", resource+" not found")
}

// Unauthorized 未授权响应
func Unauthorized(c *gin.Context, message string) {
	Error(c, http.StatusUnauthorized, "Unauthorized", message)
}

// Forbidden 禁止访问响应
func Forbidden(c *gin.Context, message string) {
	Error(c, http.StatusForbidden, "Forbidden", message)
}

// InternalError 内部服务器错误响应
func InternalError(c *gin.Context, message string) {
	Error(c, http.StatusInternalServerError, "Internal server error", message)
}

// Conflict 冲突响应
func Conflict(c *gin.Context, message string) {
	Error(c, http.StatusConflict, "Conflict", message)
}

// BadRequest 错误请求响应
func BadRequest(c *gin.Context, message string) {
	Error(c, http.StatusBadRequest, "Bad request", message)
}

// Created 创建成功响应
func Created(c *gin.Context, message string, data interface{}) {
	Success(c, http.StatusCreated, message, data)
}

// OK 成功响应
func OK(c *gin.Context, message string, data interface{}) {
	Success(c, http.StatusOK, message, data)
}

// NoContent 无内容响应
func NoContent(c *gin.Context, message string) {
	Success(c, http.StatusNoContent, message, nil)
}

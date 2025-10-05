package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type baseResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
}

// OK writes a 200 response with baseResponse and data
func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, baseResponse{Success: true, Data: data})
}

// Created writes a 201 response with baseResponse and data
func Created(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, baseResponse{Success: true, Data: data})
}

// Error writes an error baseResponse with the given status and message
func Error(c *gin.Context, status int, message string) {
	c.JSON(status, baseResponse{Success: false, Message: message})
}

// Convenience wrappers
func BadRequest(c *gin.Context, message string) { Error(c, http.StatusBadRequest, message) }
func NotFound(c *gin.Context, message string)   { Error(c, http.StatusNotFound, message) }
func Conflict(c *gin.Context, message string)   { Error(c, http.StatusConflict, message) }
func Internal(c *gin.Context, message string)   { Error(c, http.StatusInternalServerError, message) }

package response

import (
	"github.com/gin-gonic/gin"
)

type Meta map[string]any

type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data"`
	Meta    any    `json:"meta,omitempty"`
}

func Success(c *gin.Context, status int, message string, data any) {
	c.JSON(status, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func SuccessWithMeta(c *gin.Context, status int, message string, data any, meta Meta) {
	c.JSON(status, Response{
		Success: true,
		Message: message,
		Data:    data,
		Meta:    meta,
	})
}

type ErrorBody struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

func Error(c *gin.Context, status int, message string) {
	c.AbortWithStatusJSON(status, ErrorBody{
		Success: false,
		Message: message,
	})
}

package model

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

const (
	CodeOK                       = 0
	CodeInvalidParams            = 40001
	CodeInvalidDateFormat        = 40002
	CodeContentNotFound          = 40003
	CodeUnauthorized             = 40004
	CodeForbidden                = 40005
	CodeDuplicateDate            = 40006
	CodeDuplicateHost            = 40007
	CodeSceneDomainNotFound      = 40009
	CodeDuplicateScenePageConfig = 40010
	CodeScenePageConfigNotFound  = 40011
	CodeInternalServerError      = 50000
)

func SuccessResponse(data any) Response {
	return Response{
		Code:    CodeOK,
		Message: "ok",
		Data:    data,
	}
}

func ErrorResponse(code int, message string) Response {
	return Response{
		Code:    code,
		Message: message,
		Data:    nil,
	}
}

func WriteSuccess(c *gin.Context, status int, data any) {
	c.JSON(status, SuccessResponse(data))
}

func WriteError(c *gin.Context, status int, code int, message string) {
	c.JSON(status, ErrorResponse(code, message))
}

func WriteOK(c *gin.Context, data any) {
	WriteSuccess(c, http.StatusOK, data)
}

package model

type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

const (
	CodeOK                  = 0
	CodeInvalidParams       = 40001
	CodeInvalidDateFormat   = 40002
	CodeContentNotFound     = 40003
	CodeUnauthorized        = 40004
	CodeForbidden           = 40005
	CodeDuplicateDate       = 40006
	CodeInternalServerError = 50000
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

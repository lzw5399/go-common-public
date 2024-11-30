package httputil

import (
	"github.com/gin-gonic/gin"
	ferrors "github.com/lzw5399/go-common-public/library/errors"
)

type CommonHttpWrapperRsp struct {
	Error   string            `json:"error"`
	ErrCode ferrors.ErrorCode `json:"errcode"`
	Data    interface{}       `json:"data"`
	TraceId string            `json:"traceid"`
}

func MakeRsp(c *gin.Context, httpStatus int, errCode ferrors.ErrorCode, data interface{}, errCodeArgs ...interface{}) {
	if data == nil {
		data = gin.H{}
	}

	c.JSON(httpStatus, CommonHttpWrapperRsp{
		Error:   ferrors.GetErrorMessageWithLang(c, errCode, errCodeArgs...),
		ErrCode: errCode,
		Data:    data,
		TraceId: traceIdFromGinCtx(c),
	})
	c.Abort()
}

func MakeRspWithRspInfo(c *gin.Context, rspInfo *ferrors.SvrRspInfo, data interface{}) {
	if rspInfo == nil {
		rspInfo = ferrors.Ok()
	}

	if data == nil {
		data = gin.H{}
	}

	// 优先使用验证器的消息
	message := rspInfo.ValidatorRawMessage
	if message == "" {
		message = ferrors.GetErrorMessageWithLang(c, rspInfo.ErrCode, rspInfo.Args...)
	}

	c.JSON(rspInfo.HttpStatus, CommonHttpWrapperRsp{
		Error:   message,
		ErrCode: rspInfo.ErrCode,
		Data:    data,
		TraceId: traceIdFromGinCtx(c),
	})
	c.Abort()
}

func traceIdFromGinCtx(c *gin.Context) string {
	traceIdObj := c.Value("traceid")
	if traceIdObj == nil {
		return ""
	}

	traceId, ok := traceIdObj.(string)
	if !ok {
		return ""
	}
	return traceId
}

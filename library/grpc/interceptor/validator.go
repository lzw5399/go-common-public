package interceptor

import (
	"context"
	"net/http"
	"regexp"

	ferrors "github.com/lzw5399/go-common-public/library/errors"
	"google.golang.org/grpc"
)

var (
	builtinErrMsg = map[*regexp.Regexp]ferrors.ErrorCode{
		stringEmptyRegex: ferrors.ECODE_PARAM_STRING_EMPTY_ERR,     // 字符串不为空
		inEnumRegex:      ferrors.ECODE_PARAM_NOT_IN_ENUM_ERR,      // 字段必须是一个合法的枚举值
		greaterThanRegex: ferrors.ECODE_PARAM_NOT_GREATER_THAN_ERR, // 整数必须大于
		lessThanRegex:    ferrors.ECODE_PARAM_NOT_LESS_THAN_ERR,    // 整数必须小于
	}
)

var (
	stringEmptyRegex = regexp.MustCompile(`invalid field (\w+): value '(.*?)' must not be an empty string`)
	inEnumRegex      = regexp.MustCompile(`invalid field (\w+): value '(\d+)' must be a valid (\w+) field`)
	greaterThanRegex = regexp.MustCompile(`invalid field (\w+): value '(-?\d+)' must be greater than '(\d+)'`)
	lessThanRegex    = regexp.MustCompile(`invalid field (\w+): value '(-?\d+)' must be less than '(\d+)'`)
)

func ValidatorInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if err := validate(req); err != nil {
			return nil, err
		}
		return handler(ctx, req)
	}
}

type validator interface {
	Validate() error
}

func validate(req interface{}) error {
	switch v := req.(type) {
	case validator:
		if err := v.Validate(); err != nil {
			return getErrorCodeByErr(err)
		}
	}
	return nil
}

func getErrorCodeByErr(err error) *ferrors.SvrRspInfo {
	if err == nil {
		return ferrors.Ok()
	}

	errMsg := err.Error()

	// 匹配内置的一些错误，返回更加友好的错误信息
	for regex, errCode := range builtinErrMsg {
		matches := regex.FindStringSubmatch(errMsg)
		if len(matches) <= 1 {
			continue
		}

		fieldName := matches[1]
		if fieldName == "" {
			break
		}

		switch errCode {
		case ferrors.ECODE_PARAM_NOT_GREATER_THAN_ERR, ferrors.ECODE_PARAM_NOT_LESS_THAN_ERR:
			return ferrors.New(http.StatusBadRequest, errCode, fieldName, matches[3])
		default:
			return ferrors.New(http.StatusBadRequest, errCode, fieldName)
		}
	}

	// 都没找到直接返回默认的错误信息
	return &ferrors.SvrRspInfo{
		HttpStatus:          http.StatusBadRequest,
		ErrCode:             ferrors.ECODE_PARAM_ERR,
		ValidatorRawMessage: errMsg,
	}
}

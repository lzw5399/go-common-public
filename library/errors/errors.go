package ferrors

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/lzw5399/go-common-public/library/i18n"
)

var (
	errorCodeMap = map[ErrorCode]map[i18n.Lang]string{
		OK: {
			i18n.LangZh:   "",
			i18n.LangEn:   "",
			i18n.LangZhHk: "",
		},
		ECODE_SERVER_ERR: {
			i18n.LangZh:   "服务错误",
			i18n.LangEn:   "Internal Server Error",
			i18n.LangZhHk: "服務錯誤",
		},
		ECODE_PARAM_ERR: {
			i18n.LangZh:   "参数错误",
			i18n.LangEn:   "Param Error",
			i18n.LangZhHk: "參數錯誤",
		},
		ECODE_FORBIDDEN: {
			i18n.LangZh:   "当前账户没有权限访问或操作",
			i18n.LangEn:   "Current account has no permission to access or operate",
			i18n.LangZhHk: "當前賬戶沒有權限訪問或操作",
		},
		ECODE_UNAUTHORIZED: {
			i18n.LangZh:   "会话未认证",
			i18n.LangEn:   "Unauthorized",
			i18n.LangZhHk: "會話未認證",
		},
		ECODE_PARAM_STRING_EMPTY_ERR: {
			i18n.LangZh:   "字段: %s 的值不可为空",
			i18n.LangEn:   "field: %s cannot be empty",
			i18n.LangZhHk: "字段: %s 的值不可為空",
		},
		ECODE_PARAM_NOT_IN_ENUM_ERR: {
			i18n.LangZh:   "字段: %s 的值必须是一个合法的枚举值",
			i18n.LangEn:   "field: %s must be a valid enum value",
			i18n.LangZhHk: "字段: %s 的值必須是一個合法的枚舉值",
		},
		ECODE_PARAM_NOT_GREATER_THAN_ERR: {
			i18n.LangZh:   "字段: %s 的值必须大于 %s",
			i18n.LangEn:   "field: %s must greater than %s",
			i18n.LangZhHk: "字段: %s 的值必須大於 %s",
		},
		ECODE_PARAM_NOT_LESS_THAN_ERR: {
			i18n.LangZh:   "字段: %s 的值必须小于 %s",
			i18n.LangEn:   "field: %s must less than %s",
			i18n.LangZhHk: "字段: %s 的值必須小於 %s",
		},
	}
	_once sync.Once
)

var (
	grpcErrRegex = regexp.MustCompile(`\{[^\{]*\}`)
	errCodeRegex = regexp.MustCompile(`\{error_code_msg::([^}]+)\}`)
)

type ErrorCode string

const (
	OK                 ErrorCode = "OK"
	ECODE_SERVER_ERR   ErrorCode = "ECODE_SERVER_ERR"
	ECODE_PARAM_ERR    ErrorCode = "ECODE_PARAM_ERR"
	ECODE_FORBIDDEN    ErrorCode = "ECODE_FORBIDDEN"
	ECODE_UNAUTHORIZED ErrorCode = "ECODE_UNAUTHORIZED"

	ECODE_PARAM_STRING_EMPTY_ERR     ErrorCode = "ECODE_PARAM_STRING_EMPTY_ERR"
	ECODE_PARAM_NOT_IN_ENUM_ERR      ErrorCode = "ECODE_PARAM_NOT_IN_ENUM_ERR"
	ECODE_PARAM_NOT_GREATER_THAN_ERR ErrorCode = "ECODE_PARAM_NOT_GREATER_THAN_ERR"
	ECODE_PARAM_NOT_LESS_THAN_ERR    ErrorCode = "ECODE_PARAM_NOT_LESS_THAN_ERR"
)

func RegisterErrorMap(m map[ErrorCode]map[i18n.Lang]string) {
	_once.Do(func() {
		for k, v := range m {
			for langK, langV := range v {
				needReplace, resultStr := ReplaceRegexString(langV, langK)
				if needReplace {
					v[langK] = resultStr
				}
			}
			errorCodeMap[k] = v
		}
	})
}

type SvrRspInfo struct {
	HttpStatus          int           `json:"httpStatus"`
	ErrCode             ErrorCode     `json:"errCode"`
	Args                []interface{} `json:"args"`
	ValidatorRawMessage string        `json:"-"`
}

func New(status int, errcode ErrorCode, args ...interface{}) *SvrRspInfo {
	return &SvrRspInfo{
		HttpStatus: status,
		ErrCode:    errcode,
		Args:       args,
	}
}

func Ok() *SvrRspInfo {
	return &SvrRspInfo{
		HttpStatus: http.StatusOK,
		ErrCode:    OK,
	}
}

func InternalServerError() *SvrRspInfo {
	return &SvrRspInfo{
		HttpStatus: http.StatusInternalServerError,
		ErrCode:    ECODE_SERVER_ERR,
	}
}

func Forbidden() *SvrRspInfo {
	return &SvrRspInfo{
		HttpStatus: http.StatusForbidden,
		ErrCode:    ECODE_FORBIDDEN,
	}
}

func Unauthorized() *SvrRspInfo {
	return &SvrRspInfo{
		HttpStatus: http.StatusUnauthorized,
		ErrCode:    ECODE_UNAUTHORIZED,
	}
}

// SetStatus 不再建议使用。建议每次直接 ferrors.New 来声明新的 *SvrRspInfo
func (s *SvrRspInfo) SetStatus(status int, errcode ErrorCode, args ...interface{}) *SvrRspInfo {
	s.HttpStatus = status
	s.ErrCode = errcode
	s.Args = args
	return s
}

func (s *SvrRspInfo) Valid() bool {
	if s == nil {
		return true
	}

	return s.HttpStatus == http.StatusOK
}

func (s *SvrRspInfo) Error() string {
	raw, _ := json.Marshal(s)
	return string(raw)
}

func (s *SvrRspInfo) String() string {
	return fmt.Sprintf("%s:%s", s.ErrCode, errorCodeMap[s.ErrCode])
}

func GetErrorMessageWithLang(c *gin.Context, errCode ErrorCode, args ...interface{}) string {
	langStr, ok := c.Value("lang").(string)
	if !ok {
		langStr = "zh"
	}
	lang := i18n.Lang(langStr)

	errMsg := errorCodeMap[errCode]
	errMsgWithLang := errMsg[lang]
	return fmt.Sprintf(errMsgWithLang, args...)
}

func ExtractSvrRspInfo(err error) *SvrRspInfo {
	if err == nil {
		return Ok()
	}

	// 直接能转换成SvrRspInfo的，直接返回
	if rspInfo, ok := err.(*SvrRspInfo); ok {
		return rspInfo
	}

	// 无法直接转换的，可能是grpc返回的错误，消息体里面包含了错误信息
	var rspInfo SvrRspInfo
	jsonStr := grpcErrRegex.FindString(err.Error())
	if jsonStr == "" {
		return Ok().SetStatus(http.StatusInternalServerError, ECODE_SERVER_ERR)
	}

	unmarshalErr := json.Unmarshal([]byte(jsonStr), &rspInfo)
	if unmarshalErr != nil {
		return Ok().SetStatus(http.StatusInternalServerError, ECODE_SERVER_ERR)
	}

	return &rspInfo
}

func ReplaceRegexString(str string, lang i18n.Lang) (bool, string) {
	if errCodeRegex.MatchString(str) {
		replaceFunc := func(s string) string {
			match := errCodeRegex.FindStringSubmatch(s)
			if len(match) > 1 {
				i18nKey := fmt.Sprintf("error_code_msg::%s", match[1])
				i18nFiled := i18n.T(lang, i18n.MessageCode(i18nKey))
				return i18nFiled
			}
			return s
		}
		// 替换字符串
		result := errCodeRegex.ReplaceAllStringFunc(str, replaceFunc)
		return true, result
	}
	return false, ""
}

package fcontext

import (
	"context"
	"strings"

	"github.com/google/uuid"

	fconfig "github.com/lzw5399/go-common-public/library/config"
	"github.com/lzw5399/go-common-public/library/i18n"
)

type Platform string

const (
	PLATFORM_UNKNOWN Platform = "unknown"
	PLATFORM_CLIENT  Platform = "client" // 用户端
	PLATFORM_OPS     Platform = "ops"    // 运营端
)

type UserInfo struct {
	PlatForm Platform `json:"platform"` // 当前用户是「用户端」还是「运营端」
	IsAdmin  bool     `json:"isAdmin"`  // 是否是管理员。目前暂时只有「运营端」有管理员
	UserId   int64    `json:"userId"`   // dev_account 或者 oper_account 表的 id
	OrgId    int64    `json:"orgId"`    // dev_organ 或者 oper_organ 表的 id
	MemberId int64    `json:"memberId"` // member 表的 id
}

func UserInfoWithContext(ctx context.Context, userInfo *UserInfo) context.Context {
	return context.WithValue(ctx, HeaderUserInfo, userInfo)
}

func UserInfoFromContext(ctx context.Context) *UserInfo {
	userInfoObj := ctx.Value(HeaderUserInfo)
	if userInfoObj == nil {
		return nil
	}

	userInfo, ok := userInfoObj.(*UserInfo)
	if !ok {
		return nil
	}

	return userInfo
}

func LangFromContext(ctx context.Context) i18n.Lang {
	langObj := ctx.Value(i18n.HeaderLang)
	if langObj == nil {
		return i18n.Lang(fconfig.DefaultConfig.DefaultLang)
	}

	lang, ok := langObj.(string)
	if !ok {
		return i18n.Lang(fconfig.DefaultConfig.DefaultLang)
	}

	return i18n.Lang(lang)
}

func TraceIdFromContext(ctx context.Context) string {
	traceIdObj := ctx.Value(HeaderTraceId)
	if traceIdObj == nil {
		return ""
	}

	traceId, ok := traceIdObj.(string)
	if !ok {
		return ""
	}

	return traceId
}

func TraceIdWithContext(ctx context.Context, traceId string) context.Context {
	return context.WithValue(ctx, HeaderTraceId, traceId)
}

const defaultCaller = "http"

func CallerFromContext(ctx context.Context) string {
	callerObj := ctx.Value(HeaderCaller)
	if callerObj == nil {
		return defaultCaller
	}

	caller, ok := callerObj.(string)
	if !ok {
		return defaultCaller
	}

	return caller
}

func ClientIpFromContext(ctx context.Context) string {
	clientIpObj := ctx.Value(HeaderClientIp)
	if clientIpObj == nil {
		return ""
	}

	clientIp, ok := clientIpObj.(string)
	if !ok {
		return ""
	}

	return clientIp
}

func PastCallerFromContext(ctx context.Context) string {
	pastCallerObj := ctx.Value(HeaderPastCaller)
	if pastCallerObj == nil {
		return ""
	}

	pastCaller, ok := pastCallerObj.(string)
	if !ok {
		return ""
	}

	return pastCaller
}

func AuthorizationFromContext(ctx context.Context) string {
	authObj := ctx.Value(HeaderAuthorization)
	if authObj == nil {
		return ""
	}

	auth, ok := authObj.(string)
	if !ok {
		return ""
	}

	return auth
}

func HttpEndpointFromContext(ctx context.Context) string {
	httpUrlObj := ctx.Value(HeaderHttpEndpoint)
	if httpUrlObj == nil {
		return ""
	}

	httpUrl, ok := httpUrlObj.(string)
	if !ok {
		return ""
	}

	return httpUrl
}

func IsFromOpenApiWithContext(ctx context.Context, fromOpenApi bool) context.Context {
	fromOpenApiStr := "0"
	if fromOpenApi {
		fromOpenApiStr = "1"
	}
	return context.WithValue(ctx, HeaderFromOpenApi, fromOpenApiStr)
}

func IsFromOpenApiFromContext(ctx context.Context) bool {
	fromOpenApiObj := ctx.Value(HeaderFromOpenApi)
	if fromOpenApiObj == nil {
		return false
	}

	fromOpenApi, ok := fromOpenApiObj.(string)
	if !ok {
		return false
	}

	return fromOpenApi == "1"
}

func RefererWithContext(ctx context.Context, referer string) context.Context {
	return context.WithValue(ctx, HeaderReferer, referer)
}

func RefererFromContext(ctx context.Context) string {
	refererObj := ctx.Value(HeaderReferer)
	if refererObj == nil {
		return ""
	}

	referer, ok := refererObj.(string)
	if !ok {
		return ""
	}

	return referer
}

func Background() context.Context {
	ctx := context.Background()
	ctx = TraceIdWithContext(ctx, strings.ReplaceAll(uuid.New().String(), "-", ""))
	return ctx
}

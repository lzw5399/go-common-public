package interceptor

import (
	"context"
	"encoding/json"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"

	fconfig "github.com/lzw5399/go-common-public/library/config"
	fcontext "github.com/lzw5399/go-common-public/library/context"
	"github.com/lzw5399/go-common-public/library/i18n"
	"github.com/lzw5399/go-common-public/library/log"
	"github.com/lzw5399/go-common-public/library/util"
)

func OutgoingMetadataInterceptor(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	pairs := metadata.Pairs()

	// caller
	caller := fconfig.DefaultConfig.ServerName
	pairs.Set(fcontext.HeaderCaller, caller)

	// past caller
	pastCaller := mergeCurrentCallerToPastCaller(ctx)
	pairs.Set(fcontext.HeaderPastCaller, pastCaller)

	// lang
	langObj := ctx.Value(i18n.HeaderLang)
	lang, ok := langObj.(string)
	if !ok || lang == "" {
		lang = fconfig.DefaultConfig.DefaultLang
	}
	pairs.Set(i18n.HeaderLang, lang)

	// accountInfo
	accountInfoObj := ctx.Value(fcontext.HeaderUserInfo)
	accountInfo, ok := accountInfoObj.(*fcontext.UserInfo)
	if ok {
		accountInfoRaw, _ := json.Marshal(accountInfo)
		pairs.Set(fcontext.HeaderUserInfo, string(accountInfoRaw))
	}

	// trace id
	traceIdObj := ctx.Value(fcontext.HeaderTraceId)
	traceId, ok := traceIdObj.(string)
	if !ok {
		traceId = util.NewUUIDString()
	}
	pairs.Set(fcontext.HeaderTraceId, traceId)

	// authorization string
	authObj := ctx.Value(fcontext.HeaderAuthorization)
	authStr, ok := authObj.(string)
	if ok {
		pairs.Set(fcontext.HeaderAuthorization, authStr)
	}

	// client ip
	clientIpObj := ctx.Value(fcontext.HeaderClientIp)
	clientIp, ok := clientIpObj.(string)
	if ok {
		pairs.Set(fcontext.HeaderClientIp, clientIp)
	}

	// http endpoint
	endpointObj := ctx.Value(fcontext.HeaderHttpEndpoint)
	endpoint, ok := endpointObj.(string)
	if ok {
		pairs.Set(fcontext.HeaderHttpEndpoint, endpoint)
	}

	// from open-api
	fromOpenApiObj := ctx.Value(fcontext.HeaderFromOpenApi)
	fromOpenApi, ok := fromOpenApiObj.(string)
	if ok {
		pairs.Set(fcontext.HeaderFromOpenApi, fromOpenApi)
	}

	// referer
	refererObj := ctx.Value(fcontext.HeaderReferer)
	referer, ok := refererObj.(string)
	if ok {
		pairs.Set(fcontext.HeaderReferer, referer)
	}

	ctx = metadata.NewOutgoingContext(ctx, pairs)

	err := invoker(ctx, method, req, reply, cc, opts...)
	return err
}

func InComingMetadataInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	// 获取metadata
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return handler(ctx, req)
	}

	// 读取 kv
	for k, v := range md {
		if len(v) == 0 {
			continue
		}

		switch k {
		case fcontext.HeaderUserInfo: // 用户信息特殊处理，需要反序列化
			var userInfo fcontext.UserInfo
			err := json.Unmarshal([]byte(v[0]), &userInfo)
			if err != nil {
				log.Errorc(ctx, "InComingMetadataInterceptor unmarshal user info failed: %s", err)
				continue
			}
			ctx = context.WithValue(ctx, k, &userInfo)

		default:
			ctx = context.WithValue(ctx, k, v[0])
		}
	}

	// 继续处理
	resp, err = handler(ctx, req)
	if resp == nil {
		resp = &emptypb.Empty{}
	}
	return resp, err
}

func mergeCurrentCallerToPastCaller(ctx context.Context) string {
	pastCallerObj := ctx.Value(fcontext.HeaderPastCaller)
	pastCaller, ok := pastCallerObj.(string)
	if !ok {
		pastCaller = ""
	}

	currentCallerObj := ctx.Value(fcontext.HeaderCaller)
	currentCaller, ok := currentCallerObj.(string)
	if !ok {
		currentCaller = ""
	}

	// 将当前caller添加到已有的 pastCaller 中
	pastArr := strings.Split(pastCaller, ",")
	pastArr = append(pastArr, currentCaller)
	pastCaller = strings.Join(pastArr, ",")

	trimmedArr := make([]string, 0, len(pastArr))
	for _, v := range pastArr {
		if v == "" {
			continue
		}
		trimmedArr = append(trimmedArr, v)
	}

	return strings.Join(trimmedArr, ",")
}

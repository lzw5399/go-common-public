package fcontext

const (
	HeaderUserInfo      = "fc-user-info"     // 当前用户信息
	HeaderTraceId       = "traceid"          // 每个请求链路唯一的标识
	HeaderCaller        = "caller"           // 用于标识调用方
	HeaderPastCaller    = "pastcaller"       // 用于标识调用链
	HeaderAuthorization = "fc-auth"          // http请求中的Authorization header 去除Bearer之后的信息
	HeaderClientIp      = "fc-client-ip"     // 客户端ip
	HeaderHttpEndpoint  = "endpoint-bin"     // http请求的api
	HeaderFromOpenApi   = "fc-from-open-api" // 是否来自于open-api的请求
	HeaderReferer       = "referer"          // referer
)

package fconfig

type DBMode string

const (
	DB_MODE_MONGO DBMode = "mongo"
	DB_MODE_MYSQL DBMode = "mysql"
	DB_MODE_DM    DBMode = "dm"
	DB_MODE_GODEN DBMode = "golden"
)

type RedisMode string

const (
	REDIS_MODE_SINGLE   RedisMode = "single"
	REDIS_MODE_SENTINEL RedisMode = "sentinel"
	REDIS_MODE_CLUSTER  RedisMode = "cluster"
)

type Edition string

const (
	EDITION_SAAS      Edition = "saas"
	EDITION_PRIVATE   Edition = "private"
	EDITION_COMMUNITY Edition = "community"
	EDITION_POC       Edition = "poc"
)

type CdnMode string

const (
	CDN_MODE_DIRECT   CdnMode = "direct"   // cdn模式：直连
	CDN_MODE_REDIRECT CdnMode = "redirect" // cdn模式：重定向，首先访问重定向地址，经过cdn前置操作比如流量统计，再重定向回cdn访问地址
)

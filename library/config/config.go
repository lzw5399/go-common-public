package fconfig

import (
	"fmt"
	"net"
	"strings"

	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
	"github.com/pkg/errors"
)

var (
	DefaultConfig Config
)

// Init
// 配置优先级 envDefault < config.conf < 真实环境变量(yaml helm等) < config.{env}.conf < quark的registry
func Init(customConfig ICustomConfig, confBaseDir string) {
	// 使用 config.conf 作为保底的配置文件
	baseConfFile := fmt.Sprintf("%s/config.conf", confBaseDir)
	_ = godotenv.Load(baseConfFile)

	// parse config.conf 保底配置文件
	if err := env.Parse(&DefaultConfig); err != nil {
		panic(errors.Wrap(err, "parse default config error"))
	}

	// 使用 config.{env}.conf 作为的覆盖base的配置文件
	envConfFile := fmt.Sprintf("%s/config.%s.conf", confBaseDir, DefaultConfig.Env)
	_ = godotenv.Overload(envConfFile)

	// parse config.{env}.conf 配置文件
	if err := env.Parse(&DefaultConfig); err != nil {
		panic(errors.Wrap(err, "parse default config error"))
	}

	if err := env.Parse(customConfig); err != nil {
		panic(errors.Wrap(err, "parse custom config error"))
	}

	// 初始化一些需要计算的字段
	initRequiredFields(&DefaultConfig)

	customConfig.SetBaseConfig(&DefaultConfig)
}

type ICustomConfig interface {
	SetBaseConfig(config *Config)
}

type Config struct {
	ServerName             string  `env:"SERVER_NAME" envDefault:"sampleName" json:"SERVER_NAME"`               // 服务名称
	Env                    string  `env:"ENV" envDefault:"fc-community" json:"ENV"`                             // 部署环境
	Edition                Edition `env:"EDITION" envDefault:"community" json:"EDITION"`                        // 产品版本: uat/private/community
	HTTPPort               string  `env:"HTTP_PORT" envDefault:"8080" json:"HTTP_PORT"`                         // 服务监听的http端口
	RouterPrefix           string  `env:"ROUTER_PREFIX" envDefault:"" json:"ROUTER_PREFIX"`                     // 额外的路由前缀
	GRPCDiscoveryMode      string  `env:"GRPC_DISCOVERY_MODE" envDefault:"registry" json:"GRPC_DISCOVERY_MODE"` // grpc服务发现的模式。可选: registry, direct
	GRPCPort               string  `env:"GRPC_PORT" envDefault:"9090" json:"GRPC_PORT"`                         // 服务监听的grpc端口
	GRPCRoundRobin         bool    `env:"GRPC_ROUND_ROBIN" envDefault:"true" json:"GRPC_ROUND_ROBIN"`           // grpc负载均衡是否开启轮询
	BallastSizeMB          int     `env:"BALLAST_SIZE_MB" envDefault:"1024" json:"BALLAST_SIZE_MB"`             // ballast size, 单位MB
	LogConfig                      // 日志配置
	StorageConfig                  // 对象存储配置
	LanguageConfig                 // 语言配置
	EncryptConfig                  // 加密相关配置
	DBConfig                       // 数据库
	RegistryConfig                 // registry 配置
	RedisConfig                    // redis配置
	MqConfig                       // 消息队列
	TraceConfig                    // 链路追踪
	MetricConfig                   // 监控
	LicenseManagerConfig           // license相关配置
	ServiceDiscoveryConfig         // 服务发现配置
	HttpSecurityConfig             // http安全配置
	CdnClientConfig                // cdn配置
	OrgDictConfig                  //org配置
}

type HttpSecurityConfig struct {
	CORSAllowOrigins             string              `env:"CORS_ALLOW_ORIGINS" envDefault:"*" json:"CORS_ALLOW_ORIGINS"`                         // 允许跨域的域名
	RefererAllowDomains          string              `env:"REFERER_ALLOW_DOMAINS" envDefault:"*" json:"REFERER_ALLOW_DOMAINS"`                   // 前端请求的referer允许的域名, 按逗号分隔
	XForwardedForAllowNetCIDR    string              `env:"X_FORWARDED_FOR_ALLOW_NET_CIDR" envDefault:"*" json:"X_FORWARDED_FOR_ALLOW_NET_CIDR"` // 允许的X-Forwarded-For的网段
	TrustedProxiesCIDR           string              `env:"TRUSTED_PROXIES_CIDR" envDefault:"*" json:"TRUSTED_PROXIES_CIDR"`                     // 信任的网关ip
	RefererAllowDomainSet        map[string]struct{} // 前端请求的referer允许的域名, 按逗号分隔
	XForwardedForAllowNetCIDRArr []*net.IPNet        // 允许的X-Forwarded-For的网段
	TrustedProxiesCIDRArr        []string
}

type ServiceDiscoveryConfig struct {
	GRPCAddrAppManager    string `json:"GRPC_ADDR_APP_MANAGER" env:"GRPC_ADDR_APP_MANAGER" envDefault:"finclip-cloud-app-manager"`          // grpc mode=registry时候，是服务名。grpc mode=direct时候，是服务名或ip:port
	GRPCAddrUserSystem    string `json:"GRPC_ADDR_USER_SYSTEM" env:"GRPC_ADDR_USER_SYSTEM" envDefault:"finclip-cloud-user-system"`          // grpc mode=registry时候，是服务名。grpc mode=direct时候，是服务名或ip:port
	GRPCAddrOperAbility   string `json:"GRPC_ADDR_OPER_ABILITY" env:"GRPC_ADDR_OPER_ABILITY" envDefault:"finclip-cloud-oper-ability"`       // grpc mode=registry时候，是服务名。grpc mode=direct时候，是服务名或ip:port
	GRPCAddrDataCenter    string `json:"GRPC_ADDR_DATA_CENTER" env:"GRPC_ADDR_DATA_CENTER" envDefault:"finclip-cloud-data-center"`          // grpc mode=registry时候，是服务名。grpc mode=direct时候，是服务名或ip:port
	GRPCAddrBillingCenter string `json:"GRPC_ADDR_BILLING_CENTER" env:"GRPC_ADDR_BILLING_CENTER" envDefault:"finclip-cloud-billing-center"` // grpc mode=registry时候，是服务名。grpc mode=direct时候，是服务名或ip:port
	GRPCAddrAIManager     string `json:"GRPC_ADDR_AI_MANAGER" env:"GRPC_ADDR_AI_MANAGER" envDefault:"finclip-cloud-ai-manager"`             // grpc mode=registry时候，是服务名。grpc mode=direct时候，是服务名或ip:port
}

type LogConfig struct {
	// 业务日志配置
	LogMode                                   string `env:"LOG_MODE" envDefault:"warn" json:"LOG_MODE"`                                                                                 // 日志级别
	EnableFileOutput                          bool   `env:"ENABLE_FILE_OUTPUT" envDefault:"false" json:"ENABLE_FILE_OUTPUT"`                                                            // 是否开启文件输出
	LogFolderPath                             string `env:"LOG_FOLDER_PATH" envDefault:"" json:"LOG_FOLDER_PATH"`                                                                       // 文件输出位置
	LogMaxSize                                int    `env:"LOG_MAX_SIZE" envDefault:"10" json:"LOG_MAX_SIZE"`                                                                           // 单文件最大容量,单位是MB
	LogMaxBackups                             int    `env:"LOG_MAX_BACKUPS" envDefault:"20" json:"LOG_MAX_BACKUPS"`                                                                     // 最大保留过期文件个数
	LogMaxSaveTime                            int    `env:"LOG_MAX_SAVE_TIME" envDefault:"10" json:"LOG_MAX_SAVE_TIME"`                                                                 // 保留过期文件的最大时间间隔,单位是天
	LogCompress                               bool   `env:"LOG_COMPRESS" envDefault:"false" json:"LOG_COMPRESS"`                                                                        // 是否需要压缩滚动日志, 使用的 gzip 压缩
	ParseSvrRspInfoAndDowngrade400SerialError bool   `env:"PARSE_SVR_RSP_INFO_AND_DOWNGRADE400_SERIAL_ERROR" envDefault:"true" json:"PARSE_SVR_RSP_INFO_AND_DOWNGRADE400_SERIAL_ERROR"` // 是否解析服务端返回的错误信息, 如果是 error 的 400系列code(400<=code<500), 则降级为warn级别

	// Gorm日志配置
	GormLogMode string `env:"GORM_LOG_MODE" envDefault:"warn" json:"GORM_LOG_MODE"`  // gorm日志级别, info 会打印输出sql。可选 silent, error, warn, info
	GormLogJson bool   `env:"GORM_LOG_JSON" envDefault:"false" json:"GORM_LOG_JSON"` // gorm日志是否输出json

	// Gin日志配置
	GinMode string `env:"GIN_MODE" envDefault:"release" json:"GIN_MODE"` // gin模式
}

type StorageConfig struct {
	StorageMode string `env:"STORAGE_MODE" envDefault:"minio" json:"STORAGE_MODE"` // 后端的存储实现。 可选类型 minio, ali_oss, aws_s3, tencent_cos, disk

	// 对外提供服务的基础配置
	StorageApiKey string `env:"STORAGE_API_KEY" envDefault:"234CB575090D52CE2DF0E71592850A1B99433CCB7523A3D349E06F73FEC80EBA" json:"STORAGE_API_KEY"`

	// image
	StorageImageContentTypeWhitelistStr string `env:"STORAGE_IMAGE_CONTENT_TYPE_WHITELIST_STR" envDefault:"image/png,application/x-png,image/jpeg,image/gif,image/webp,image/bmp,image/svg+xml,image/x-icon,image/vnd.microsoft.icon" json:"STORAGE_IMAGE_CONTENT_TYPE_WHITELIST_STR"` // 支持的图片上传后缀名
	StorageImageExtensionWhitelistStr   string `env:"STORAGE_IMAGE_EXTENSION_WHITELIST_STR" envDefault:"png,jpeg,jpg,gif,webp,bmp,ico,svg" json:"STORAGE_IMAGE_EXTENSION_WHITELIST_STR"`                                                                                               // 支持的图片上传后缀名
	// doc
	StorageDocContentTypeWhitelistStr string `env:"STORAGE_DOC_CONTENT_TYPE_WHITELIST_STR" envDefault:"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet,application/vnd.openxmlformats-officedocument.wordprocessingml.document,application/vnd.ms-excel,application/msword,text/plain,application/pdf,application/x-ole-storage" json:"STORAGE_DOC_CONTENT_TYPE_WHITELIST_STR"` // 支持的文档上传后缀名
	StorageDocExtensionWhitelistStr   string `env:"STORAGE_DOC_EXTENSION_WHITELIST_STR" envDefault:"xls,xlsx,doc,docx,txt,pdf" json:"STORAGE_DOC_EXTENSION_WHITELIST_STR"`                                                                                                                                                                                                                        // 支持的文档上传后缀名
	// zip
	StorageZipContentTypeWhitelistStr string `env:"STORAGE_ZIP_CONTENT_TYPE_WHITELIST_STR" envDefault:"application/x-zip-compressed,application/zip,application/zstd,application/octet-stream" json:"STORAGE_ZIP_CONTENT_TYPE_WHITELIST_STR"` // 支持的压缩包上传后缀名
	StorageZipExtensionWhitelistStr   string `env:"STORAGE_ZIP_EXTENSION_WHITELIST_STR" envDefault:"zip,zstd,ftpkg" json:"STORAGE_ZIP_EXTENSION_WHITELIST_STR"`                                                                               // 支持的压缩包上传后缀名
	// others 剩余其他的
	StorageOthersContentTypeWhitelistStr string `env:"STORAGE_OTHERS_CONTENT_TYPE_WHITELIST_STR" envDefault:"application/octet-stream" json:"STORAGE_OTHERS_CONTENT_TYPE_WHITELIST_STR"` // 支持的其他文件上传后缀名
	StorageOthersExtensionWhitelistStr   string `env:"STORAGE_OTHERS_EXTENSION_WHITELIST_STR" envDefault:"ipa,apk" json:"STORAGE_OTHERS_EXTENSION_WHITELIST_STR"`                        // 支持的其他文件上传后缀名

	// 通用配置
	StorageEndpoint   string `env:"STORAGE_ENDPOINT" envDefault:"" json:"STORAGE_ENDPOINT"`              // 存储服务地址
	StorageAccessKey  string `env:"STORAGE_ACCESS_KEY" envDefault:"" json:"STORAGE_ACCESS_KEY"`          // 存储服务access key
	StorageSecretKey  string `env:"STORAGE_SECRET_KEY" envDefault:"" json:"STORAGE_SECRET_KEY"`          // 存储服务secret key
	StorageBucketName string `env:"STORAGE_BUCKET_NAME" envDefault:"finclip" json:"STORAGE_BUCKET_NAME"` // 存储服务bucket name
	StorageUploadPath string `env:"STORAGE_UPLOAD_PATH" envDefault:"finclip" json:"STORAGE_UPLOAD_PATH"` // 上传路径

	// s3(以及兼容s3协议的)专有配置
	StorageS3ForcePath bool   `env:"STORAGE_S3_FORCE_PATH" envDefault:"true" json:"STORAGE_S3_FORCE_PATH"`
	StorageS3ObjectAcl string `env:"STORAGE_S3_OBJECT_ACL" envDefault:"" json:"STORAGE_S3_OBJECT_ACL"`
	StorageS3Region    string `env:"STORAGE_S3_REGION" envDefault:"cn-northwest-1" json:"STORAGE_S3_REGION"`
	StorageS3UseSSL    bool   `env:"STORAGE_S3_USE_SSL" envDefault:"false" json:"STORAGE_S3_USE_SSL"`

	// nas 专有配置
	StorageNasDiskBasePath string `env:"STORAGE_NAS_DISK_BASE_PATH" envDefault:"/tmp/netdisk/" json:"STORAGE_NAS_DISK_BASE_PATH"` // nas文件存储路径
}

type LanguageConfig struct {
	DefaultLang string `env:"DEFAULT_LANG" envDefault:"zh" json:"DEFAULT_LANG"`    // 默认语言
	SupportLang string `env:"SUPPORT_LANG" envDefault:"zh,en" json:"SUPPORT_LANG"` // 支持的语言
}

// EncryptConfig 加解密配置
type EncryptConfig struct {
	PwdHashType              string `env:"PWD_HASH_TYPE" envDefault:"sha256" json:"PWD_HASH_TYPE"`                                         // 密码hash类型。 可选类型 sha256, sm3
	CommonEncryptType        string `env:"COMMON_ENCRYPT_TYPE" envDefault:"des" json:"COMMON_ENCRYPT_TYPE"`                                // 通用敏感字段加密。 可选类型 des, sm4
	CommonEncryptDESCryptKey string `env:"COMMON_ENCRYPT_DES_CRYPT_KEY" envDefault:"w$D5%8x@" json:"COMMON_ENCRYPT_DES_CRYPT_KEY"`         // 通用敏感字段加密。
	CommonEncryptSm4CryptKey string `env:"COMMON_ENCRYPT_SM4_CRYPT_KEY" envDefault:"finclip9876cloud" json:"COMMON_ENCRYPT_SM4_CRYPT_KEY"` // 通用敏感字段加密。
}

type DBConfig struct {
	DBMode          DBMode `env:"DB_MODE" envDefault:"mysql" json:"DB_MODE"`                                                                                                           // 数据库模式
	MysqlURL        string `env:"MYSQL_URL" envDefault:"root:bcMy1GnlOTJZTnR3K/QYhvhZ2VwOJaYyO9ZRpFpjZkg=@tcp(192.168.0.11:30006)/finclip-cloud-dev?charset=utf8mb4" json:"MYSQL_URL"` // mysql连接字符串
	MongoURL        string `env:"MONGO_URL" envDefault:"mongodb://localhost:27017" json:"MONGO_URL"`
	DmURL           string `env:"DM_URL" envDefault:"dm://SYSDBA:SYSDBA001@127.0.0.1:5236?autoCommit=true" json:"DM_URL"` // dm连接字符串
	MaxIdleConns    int    `env:"MAX_IDLE_CONNS" envDefault:"2" json:"MAX_IDLE_CONNS"`                                    // MaxIdleConns
	MaxOpenConns    int    `env:"MAX_OPEN_CONNS" envDefault:"100" json:"MAX_OPEN_CONNS"`                                  // MaxOpenConns
	EnableMigration bool   `env:"ENABLE_MIGRATION" envDefault:"true" json:"ENABLE_MIGRATION"`                             // 是否开启数据库迁移
}

type LicenseManagerConfig struct {
	LicenseManagerAddr string `env:"LICENSE_MANAGER_ADDR" envDefault:"http://license-manager:8080" json:"LICENSE_MANAGER_ADDR"` // license manager服务地址
}

type MqConfig struct {
	MQMode string `env:"MQ_MODE" envDefault:"n" json:"MQ_MODE"` // mq模式. 支持nats, k n:n k:k

	// n
	NUrl string `env:"N_URL" envDefault:"nats://admin:password@localhost:4222" json:"N_URL"` // n地址

	// k
	KAddr      string `env:"K_ADDR" envDefault:"kafka-service.k:9093" json:"K_ADDR"` // 地址
	KVersion   string `env:"K_VERSION" envDefault:"2.3.0" json:"K_VERSION"`          // 设定版本
	KUser      string `env:"K_USER" envDefault:"" json:"K_USER"`                     // 用户
	KPwd       string `env:"K_PWD" envDefault:"" json:"K_PWD"`                       // 密码
	KMechanism string `env:"K_MECHANISM" envDefault:"PLAIN"`
	KLog       bool   `env:"K_LOG" envDefault:"false" json:"K_LOG"`               // 日志是否输入
	KLogTopic  string `env:"K_LOG_TOPIC" envDefault:"elk-log" json:"K_LOG_TOPIC"` // 日志输入topic
}

type TraceConfig struct {
	SkyWalkingUrl        string `env:"SKYWALKING_URL" envDefault:"127.0.0.1:11800" json:"SKYWALKING_URL"` // skywalking地址
	SkyWalkingEnable     bool   `env:"SKYWALKING_ENABLE" envDefault:"false" json:"SKYWALKING_ENABLE"`     // 是否打开skywalking
	SkyWalkingPartitions uint32 `env:"SKYWALKING_PARTITIONS" envDefault:"1" json:"SKYWALKING_PARTITIONS"` // skywalking的频率
}

type RegistryConfig struct {
	RegistryAddr         string `env:"REGISTRY_ADDR" envDefault:"localhost:8500" json:"REGISTRY_ADDR"`                         // 地址
	RegistryTag          string `env:"REGISTRY_TAG" envDefault:"mop-finstore" json:"REGISTRY_TAG"`                             // 服务注册的tag
	RegistryKVConfigPath string `env:"REGISTRY_KV_CONFIG_PATH" envDefault:"finclip/config/uat" json:"REGISTRY_KV_CONFIG_PATH"` // kv path, 拼接得到public和服务配置的key
}

type RedisConfig struct {
	RedisMode RedisMode `env:"REDIS_MODE" envDefault:"single" json:"REDIS_MODE"` // redis模式
	// 集群，单例模式
	RedisAddr     string `env:"REDIS_ADDR" envDefault:"redis:6379" json:"REDIS_ADDR"` // redis地址，集群或单例地址
	RedisPassword string `env:"REDIS_PASSWORD" envDefault:"" json:"REDIS_PASSWORD"`   // redis密码
	// 哨兵模式
	RedisSentinelAddr     string `env:"REDIS_SENTINEL_ADDR" envDefault:"" json:"REDIS_SENTINEL_ADDR"`         // redis哨兵地址
	RedisMasterName       string `env:"REDIS_MASTER_NAME" envDefault:"" json:"REDIS_MASTER_NAME"`             // redis哨兵master名称
	RedisDatabase         int    `env:"REDIS_INDEX" envDefault:"11" json:"REDIS_INDEX"`                       // redis哨兵数据库idx
	RedisSentinelPassword string `env:"REDIS_SENTINEL_PASSWORD" envDefault:"" json:"REDIS_SENTINEL_PASSWORD"` // redis哨兵密码
	RedisSentinelMode     string `env:"SENTINEL_MODE" envDefault:"sentinelNormal" json:"SENTINEL_MODE"`       // redis哨兵模式
}

type MetricConfig struct {
	MonitorPort        string `env:"MONITOR_PORT" envDefault:"9092" json:"MONITOR_PORT"`                  // 监控端口
	OpenMonitor        bool   `env:"OPEN_MONITOR" envDefault:"false" json:"OPEN_MONITOR"`                 // 是否打开监控
	SlowSqlMillSeconds int64  `env:"SLOW_SQL_MILL_SECONDS" envDefault:"200" json:"SLOW_SQL_MILL_SECONDS"` // 慢sql阈值, 用于告警
	PProfEnable        bool   `env:"PPROF_ENABLE" envDefault:"false" json:"PPROF_ENABLE"`                 // 是否打开pprof
}

type CdnClientConfig struct {
	CdnOpen            bool   `json:"CDN_OPEN" env:"CDN_OPEN" envDefault:"false"`                       //cdn是否开启
	CdnCachePushOpen   bool   `json:"CDN_CACHE_PUSH_OPEN" env:"CDN_CACHE_PUSH_OPEN" envDefault:"false"` //cdn预热是否开启
	CdnCachePushMode   string `json:"CDN_CACHE_PUSH_MODE" env:"CDN_CACHE_PUSH_MODE" envDefault:""`      //cdn的预热模式,取值 tencent/ali
	CdnAccessKeyId     string `json:"CDN_ACCESS_KEY_ID" env:"CDN_ACCESS_KEY_ID" envDefault:""`          //cdn的AccessKeyId
	CdnAccessKeySecret string `json:"CDN_ACCESS_KEY_SECRET" env:"CDN_ACCESS_KEY_SECRET" envDefault:""`  //cdn的AccessKeySecret

	CdnDomain string  `json:"CDN_DOMAIN" env:"CDN_DOMAIN" envDefault:""`                                        //cdn提供商配置的cdn域名
	CdnMode   CdnMode `json:"CDN_MODE" env:"CDN_MODE" envDefault:"direct"`                                      //cdn模式，直连和重定向
	CdnPreUri string  `json:"CDN_PRE_URI" env:"CDN_PRE_URI" envDefault:"/api/v1/mop/runtime/download/cdn-pre/"` //cdn重定向前置步骤地址
	CdnUri    string  `json:"CDN_URI" env:"CDN_URI" envDefault:"/api/v1/mop/runtime/download/"`                 //cdn回源请求的uri
}

type OrgDictConfig struct {
	OrgDictJson string `json:"ORG_DICT_JSON" env:"ORG_DICT_JSON" envDefault:"{\"{orgField}\":{\"zh\":\"企业\",\"en\":\"Organization\",\"zh-HK\":\"企業\"}}"`
}

func (c *Config) IsSaas() bool {
	return c.Edition == EDITION_SAAS
}

func (c *Config) IsCommunity() bool {
	return c.Edition == EDITION_COMMUNITY
}

func (c *Config) IsPrivate() bool {
	return c.Edition == EDITION_PRIVATE
}

func (c *Config) IsPOC() bool {
	return c.Edition == EDITION_POC
}

func initRequiredFields(config *Config) {
	if DefaultConfig.DefaultLang == "" {
		panic(errors.New("parse config failed. DEFAULT_LANG cannot be empty"))
	}

	// 初始化前端请求的referer允许的域名
	config.RefererAllowDomainSet = make(map[string]struct{})
	for _, v := range strings.Split(config.RefererAllowDomains, ",") {
		config.RefererAllowDomainSet[strings.TrimSpace(strings.ToLower(v))] = struct{}{}
	}

	// 初始化X-Forwarded-For的网段
	config.XForwardedForAllowNetCIDRArr = make([]*net.IPNet, 0)
	if config.HttpSecurityConfig.XForwardedForAllowNetCIDR != "*" &&
		config.HttpSecurityConfig.XForwardedForAllowNetCIDR != "" {
		for _, v := range strings.Split(config.HttpSecurityConfig.XForwardedForAllowNetCIDR, ",") {
			_, ipNet, err := net.ParseCIDR(v)
			if err != nil {
				panic(errors.Wrap(err, "parse X-Forwarded-For allow net error"))
			}
			config.XForwardedForAllowNetCIDRArr = append(config.XForwardedForAllowNetCIDRArr, ipNet)
		}
	}

	config.TrustedProxiesCIDRArr = make([]string, 0)
	if config.TrustedProxiesCIDR != "*" &&
		config.TrustedProxiesCIDR != "" {
		for _, v := range strings.Split(config.HttpSecurityConfig.TrustedProxiesCIDR, ",") {
			config.TrustedProxiesCIDRArr = append(config.TrustedProxiesCIDRArr, v)
		}
	}
}

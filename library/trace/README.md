# 使用步骤

0. 注意事项：当发布改apm新版本时，请务必修改好build.go文件下的version变量，比如:

```
var (
	lock    = &sync.Mutex{}
	client  *Client
	version = "v0.2.3"
)
```

1. 将需要使用该apm工具的项目移动到$GOPATH/src目录下。

2. 在需要使用apm工具的项目目录下执行以下命令下载依赖：

```
go get git.finogeeks.club/base/apm
go get git.finogeeks.club/base/apm@v0.2.3
go get github.com/SkyAPM/go2sky-plugins/gin/v3
go get github.com/Shopify/sarama@v1.24.1
go get golang.org/x/text@v0.3.4
```

3. 使用以下命令会将$GOPATH下的mod同步到vendor目录下。

```
go mod vendor

注意：在执行"go mod vendor"后，记得用"git add vendor/"将依赖的文件添加到git里，否则drone构建时会缺失部分文件。
```

4. 如果需要更新apm依赖，就执行以下命令：

```
go get git.finogeeks.club/base/apm@{{git-tag}}
go mod vendor
或者：
go get git.finogeeks.club/base/apm@{{git-sha}}
go mod vendor

比如：
go get git.finogeeks.club/base/apm@v0.2.3
go mod vendor
或者：
go get git.finogeeks.club/base/apm@955a98a3c1
go mod vendor
```

# 使用样例

## 特别注意项

使用apm相关API时，不要在任何地方使用`fmt.Sprint`等字符串操作的函数，这样会非常非常影响性能！！！

## 初始化
```
apm.BuildApmClient(apm.CreateBuild(cfg.SkyWalkingUrl, cfg.ServerName, cfg.SkyWalkingPartitions, cfg.SkyWalkingEnable))
```

## no trace,使用这个方法创建出来的ctx不会被埋点
```
ctx = ApmClient().NoTraceContext(ctx)
```

## 通用exitSpan
```
eSpan := trace.ApmClient().CreateExitSpan(ctx, operationName, peer, injector)
defer eSpan.End()
```

## 通用entrySpan
```
span, ctx := trace.ApmClient().CreateEntrySpan(ctx, operationName, extractor)
defer span.End()
```

## 通用localSpan
```
span, ctx := trace.ApmClient().CreateLocalSpan(ctx, opts...)
span.SetOperationName(operationName)
defer span.End()

比如：
span, ctx := trace.ApmClient().CreateLocalSpan(context.Background())
span.SetOperationName("updateLicenseValid")
defer span.End()
```

## HTTP示例

### 注入gin middleware
```
trace.ApmClient().InjectHttpMiddleware(g)
```

### exitSpan
```
span := trace.ApmClient().CreateHttpExitSpanWithUrl(ctx, req, url)
defer span.End()
或者
span := trace.ApmClient().CreateHttpExitSpanWithUrlAndInjector(ctx, http.MethodGet, url, func(header string) error {
    headers[propagation.Header] = header
    return nil
})
defer span.End()
```

## gRPC示例

### entrySpan
```
span, ctx := trace.ApmClient().CreateGRpcEntrySpan(ctx, method)
defer span.End()
```

### exitSpan
```
span := trace.ApmClient().CreateGRpcExitSpan(ctx, operationName, method, thirdService)
defer span.End()
```

## MQ_K示例

### 配置k consumer

为使使用的k客户端库sarama支持header，需要在consumer端配置k集群版本，具体如下：

```
version, e := sarama.ParseKafkaVersion("2.3.0")
if e != nil {
    panic(err)
}
// 消费者共同配置
consumerConfig = cluster.NewConfig()
consumerConfig.Version = version
```

### entrySpan
```
span, ctx := trace.ApmClient().CreateKEntrySpan(ctx, topic, method, msg)
defer span.End()
```

### exitSpan
```
span := trace.ApmClient().CreateKExitSpan(ctx, topic, method, address, msg)
defer span.End()

span.Error(time.Now(), "topic", topic, "error", err.Error())
```

## MongoDB示例

### exitSpan
```
span := trace.ApmClient().CreateMongoExitSpan(ctx, action, address, dbName)
defer span.End()
span.Log(time.Now(), "collection", USER_REGION_C_NAME, "method", "Add")
span.Error(time.Now(), "collection", t.CollName, "method", "Find.One", "error", err.Error())

具体示例：
span := trace.ApmClient().CreateMongoExitSpan(ctx, "Table.Count", config.Cfg.MongoURL, t.dbName)
defer span.End()
span.Log(time.Now(), "collection", USER_REGION_C_NAME, "method", "Add")
```

## MySql示例

### exitSpan
```
span := trace.ApmClient().CreateMySqlExitSpan(ctx, action, address, dbName)
defer span.End()
span.Log(time.Now(), "collection", USER_REGION_C_NAME, "method", "Add")
span.Error(time.Now(), "collection", t.CollName, "method", "Find.One", "error", err.Error())

具体示例：
span := trace.ApmClient().CreateMySqlExitSpan(ctx, "Table.Count", config.Cfg.MySqlURL, t.dbName)
defer span.End()
span.Log(time.Now(), "collection", USER_REGION_C_NAME, "method", "Add")
```

## ElasticSearch示例

### exitSpan
```
span := trace.ApmClient().CreateESExitSpan(ctx, action, address)
defer span.End()
span.Log(time.Now(), "ReportMsg.InsertMany start")
span.Log(time.Now(), "index", config.Cfg.PrivateDataReportIndex, "method", "INSERT_EVENT")
span.Error(time.Now(), "index", config.Cfg.PrivateDataReportIndex, "method", "INSERT_EVENT", "error", err.Error())
```

## Redis示例

### exitSpan
```
_, _, dbName := apm.ParseURL(config.Cfg.RedisUrl)
span := trace.ApmClient().CreateRedisExitSpan(c.Request.Context(), "TestYace", config.Cfg.RedisUrl, dbName)
defer span.End()
span.Log(time.Now(), "collection", USER_REGION_C_NAME, "method", "Add")
```

### redis辅助工具

目前mop项目中使用了多种redis部署方案，可以使用类似于以下的方式处理：

```
span := createRedisSpan(ctx, "RedisSet")
span.Log(time.Now(), "key", key)
defer span.End()

func createRedisSpan(ctx context.Context, action string) go2sky.Span {
	switch config.Cfg.RedisMode {
	case MODE_SENTINEL_V2:
		span := trace.ApmClient().CreateRedisExitSpan(ctx, action,
			config.Cfg.RedisSentinelAddr, "")
		span.Log(time.Now(), "Mode", config.Cfg.RedisMode)
		span.Log(time.Now(), "MasterName", config.Cfg.RedisMasterName)
		return span
	case MODE_CLUSTER:
		span := trace.ApmClient().CreateRedisExitSpan(ctx, action,
			config.Cfg.RedisAddr, "")
		span.Log(time.Now(), "Mode", config.Cfg.RedisMode)
		return span
	case MODE_CLUSTER_V2:
		span := trace.ApmClient().CreateRedisExitSpan(ctx, action,
			config.Cfg.RedisAddr, "")
		span.Log(time.Now(), "Mode", config.Cfg.RedisMode)
		return span
	default:
		if config.Cfg.RedisMode == MODE_SENTINEL {
			span := trace.ApmClient().CreateRedisExitSpan(ctx, action,
				config.Cfg.RedisSentinelAddr, "")
			span.Log(time.Now(), "Mode", config.Cfg.RedisMode)
			span.Log(time.Now(), "MasterName", config.Cfg.RedisMasterName)
			return span
		} else {
			span := trace.ApmClient().CreateRedisExitSpan(ctx, action,
				config.Cfg.RedisAddr, "")
			span.Log(time.Now(), "Mode", config.Cfg.RedisMode)
			return span
		}
	}
}
```

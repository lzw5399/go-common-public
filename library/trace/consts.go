package trace

import "github.com/SkyAPM/go2sky"

const (
	TagMQType     go2sky.Tag = "mq.type"
	TagMQMethod   go2sky.Tag = "mq.method"
	TagGRpcMethod go2sky.Tag = "grpc.method"
)

const (
	MethodProduce = "produce"
	MethodConsume = "consume"
)

const (
	MQTypeK        = "k"
	MQTypeRabbitMQ = "rabbitmq"
)

const (
	DBTypePostgreSQL    = "postgresql"
	DBTypeMySQL         = "mysql"
	DBTypeMongoDB       = "mongodb"
	DBTypeElasticSearch = "elasticsearch"
	DBTypeRedis         = "redis"
)

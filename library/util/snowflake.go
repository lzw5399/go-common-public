package util

import (
	"fmt"
	"math"
	"strconv"

	fconfig "github.com/lzw5399/go-common-public/library/config"
	"github.com/yitter/idgenerator-go/idgen"
	"github.com/yitter/idgenerator-go/regworkerid"
)

type IdGenOption struct {
	WorkerIdBitLength byte
}

// InitIdGen 需要先初始化config，不然获取不到redis的配置
func InitIdGen(idOption IdGenOption) {
	if idOption.WorkerIdBitLength < 0 || idOption.WorkerIdBitLength > 15 {
		panic("snowflake WorkerIdBitLength must between [1, 15]")
	}
	workerIdBitLength := idOption.WorkerIdBitLength
	if workerIdBitLength == 0 {
		workerIdBitLength = 8
	}
	maxWorkerId := math.Pow(2, float64(workerIdBitLength)) - 1

	var masterName string
	if fconfig.DefaultConfig.RedisMode == fconfig.REDIS_MODE_SENTINEL {
		masterName = fconfig.DefaultConfig.RedisMasterName
	}

	registerConf := regworkerid.RegisterConf{
		Address:         fconfig.DefaultConfig.RedisAddr,
		Password:        fconfig.DefaultConfig.RedisPassword,
		DB:              fconfig.DefaultConfig.RedisDatabase,
		MasterName:      masterName,
		MaxWorkerId:     int32(maxWorkerId),
		MinWorkerId:     1,
		TotalCount:      0,
		LifeTimeSeconds: 15,
	}

	fmt.Printf("[idgen] conf.MaxWorkerId:%d\n", registerConf.MaxWorkerId)

	workerIdGen := regworkerid.RegisterOne(registerConf)
	fmt.Printf("[idgen] workerIdGen:%d\n", workerIdGen)
	if workerIdGen < 0 {
		panic("snowflake workerId gen fail")
	}
	workerId := uint16(workerIdGen)
	if workerId == 0 {
		panic("snowflake workerId is 0")
	}
	var options = &idgen.IdGeneratorOptions{
		Method:   1,
		WorkerId: workerId,
		//默认值： 2020-02-20 02:20:02
		BaseTime: 1582136402000,
		//8个字节，最多支持255个容器， 如果超过255个容器，需要增加位数
		WorkerIdBitLength: workerIdBitLength,
		//WorkerIdBitLength+SeqBitLength <= 22
		SeqBitLength:     6,
		MaxSeqNumber:     0,
		MinSeqNumber:     5,
		TopOverCostCount: 2000,
	}
	idgen.SetIdGenerator(options)
}

// NewSnowflakeID
// 针对大家在使用中经常出现的性能疑问，我给出以下3组最佳实践：
// ❄ 如果ID生成需求不超过5W个/s，不用修改任何配置参数
// ❄ 如果超过5W个/s，低于50W个/s，推荐修改：SeqBitLength=10
// ❄ 如果超过50W个/s，接近500W个/s，推荐修改：SeqBitLength=12
// 总之，增加 SeqBitLength 会让性能更高，但生成的 ID 会更长。
func NewSnowflakeID() string {
	newId := idgen.NextId()
	// 生成的ID是int64类型，int类型的索引效率更高，可以把这个字段对应的数据库字段改为int类型
	return strconv.FormatInt(newId, 10)
}

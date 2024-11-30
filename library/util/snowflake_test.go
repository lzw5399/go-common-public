package util

import (
	"fmt"
	"math"
	"testing"
	"time"

	fconfig "github.com/lzw5399/go-common-public/library/config"
	"github.com/yitter/idgenerator-go/idgen"
	"github.com/yitter/idgenerator-go/regworkerid"
)

// 生成10000个id需要多少时间 //1毫秒
func TestNewSnowflakeID(t *testing.T) {
	InitIdGen(IdGenOption{0})
	start := time.Now().UnixMilli()
	for i := 0; i < 10000; i++ {
		NewSnowflakeID()
	}
	fmt.Println(time.Now().UnixMilli() - start)
}

// 并发生成10000个id，不重复
func TestNewSnowflakeIDUnique(t *testing.T) {
	InitIdGen(IdGenOption{0})
	c := make(chan string, 10000)
	for i := 0; i < 10000; i++ {
		go func() {
			id := NewSnowflakeID()
			c <- id
		}()
	}
	m := map[string]int{}
ForEnd:
	for {
		select {
		case data := <-c:
			m[data] = 1
			fmt.Println(data)
			//如果有10000个不重复的id，就退出
			if len(m) == 10000 {
				fmt.Printf("m len is %d \n", len(m))
				break ForEnd
			}
		}
	}
}

// workId测试， 获取workId， 不会重复
func TestWorkId(t *testing.T) {
	for i := 0; i < 100; i++ {
		forTest(IdGenOption{10})
		time.Sleep(time.Millisecond * 2)
	}
}

// inner func, for testing
func forTest(idOption IdGenOption) {
	if idOption.WorkerIdBitLength < 0 || idOption.WorkerIdBitLength > 15 {
		panic("snowflake WorkerIdBitLength must between [1, 15]")
	}
	workerIdBitLength := idOption.WorkerIdBitLength
	if workerIdBitLength == 0 {
		workerIdBitLength = 8
	}
	MaxWorkerId := math.Pow(2, float64(workerIdBitLength)) - 1

	workerIdGen := regworkerid.RegisterOne(regworkerid.RegisterConf{
		Address:         "redis:6379",
		Password:        fconfig.DefaultConfig.RedisPassword,
		DB:              0,
		MasterName:      "",
		MaxWorkerId:     int32(MaxWorkerId),
		MinWorkerId:     1,
		TotalCount:      0,
		LifeTimeSeconds: 15,
	})
	if workerIdGen < 0 {
		panic("snowflake workerId gen fail")
	}
	workerId := uint16(workerIdGen)
	fmt.Printf("regWorkerId, workerId : %d \n", workerId)
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

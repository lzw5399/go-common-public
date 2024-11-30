package fredis

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/pkg/errors"

	fconfig "github.com/lzw5399/go-common-public/library/config"
)

func TestClusterScanDel(t *testing.T) {
	// 初始化redis
	fconfig.DefaultConfig.RedisAddr = "192.168.0.77:7000,192.168.0.77:7001,192.168.0.78:7000,192.168.0.78:7001,192.168.0.79:7000,192.168.0.79:7001"
	fconfig.DefaultConfig.RedisMode = "cluster"
	fconfig.DefaultConfig.RedisPassword = "vMIivRgl4jhe7WUEEH"
	Init()

	ctx := context.Background()
	t.Run("ClusterScanDel succees", func(t *testing.T) {
		for i := 0; i < 100; i++ {
			_, err := Set(ctx, "cluster_test_"+strconv.Itoa(i), "123", time.Hour)
			if err != nil {
				fmt.Printf("set redis error: %v", err)
				panic(errors.Wrap(err, "set redis error"))
			}
		}

		// scan del
		err := ScanDel(ctx, "cluster_test_*")
		if err != nil {
			fmt.Printf("scan del redis error: %v", err)
			panic(errors.Wrap(err, "scan del redis error"))
		}

		// 查找
		for i := 0; i < 100; i++ {
			val, err := Get(ctx, "cluster_test_"+strconv.Itoa(i))
			if err == nil || val != "" {
				fmt.Printf("get redis error, key: %s val: %s", "cluster_test_"+strconv.Itoa(i), val)
				panic(errors.Wrap(err, fmt.Sprintf("get redis error, key: %s val: %s", "cluster_test_"+strconv.Itoa(i), val)))
			}
		}
	})
}

func TestKeys(t *testing.T) {
	// 初始化redis
	fconfig.DefaultConfig.RedisAddr = "127.0.0.1:6379"
	fconfig.DefaultConfig.RedisMode = "single"
	Init()

	ctx := context.Background()
	t.Run("Keys succees", func(t *testing.T) {
		flushAll(ctx)

		Set(ctx, "test_key1", "test_value1", 0)
		Set(ctx, "test_keydjksajkdsj", "test_value1", 0)

		// arrange
		key := "test_key*"

		// act
		allKeys, err := Keys(ctx, key)

		// assert
		if err != nil {
			t.Errorf("Keys() error = %v", err)
			return
		}

		if len(allKeys) != 2 {
			t.Errorf("Keys() = %v, want %v", len(allKeys), 2)
			return
		}
	})

	t.Run("Keys succees large", func(t *testing.T) {
		flushAll(ctx)
		num := 5000
		for i := 0; i < num; i++ {
			Set(ctx, fmt.Sprintf("test_key_%d", i), "test_value1", 0)
		}

		Set(ctx, "ydjksajkdsj", "test_value1", 0)

		// arrange
		key := "test_key*"

		// act
		allKeys, err := Keys(ctx, key)

		// assert
		if err != nil {
			t.Errorf("Keys() error = %v", err)
			return
		}

		if len(allKeys) != num {
			t.Errorf("Keys() = %v, want %v", len(allKeys), num)
			return
		}
	})
}

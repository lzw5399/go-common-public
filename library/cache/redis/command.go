package fredis

import (
	"context"
	"errors"
	"sync"
	"time"

	fconfig "github.com/lzw5399/go-common-public/library/config"
	"github.com/lzw5399/go-common-public/library/util"

	goRedis "github.com/go-redis/redis/v8"

	ferrors "github.com/lzw5399/go-common-public/library/errors"
	"github.com/lzw5399/go-common-public/library/log"
)

func Client() goRedis.UniversalClient {
	return gUniClient
}

// Set Zero expiration means the key has no expiration time.
func Set(ctx context.Context, key string, value interface{}, expiration time.Duration) (string, error) {
	return gUniClient.Set(ctx, key, value, expiration).Result()
}

func Del(ctx context.Context, key ...string) (int64, error) {
	cfg := fconfig.DefaultConfig
	// 集群模式下，多个值，需要逐个删除
	if len(key) > 0 && cfg.RedisMode == fconfig.REDIS_MODE_CLUSTER {
		for _, k := range key {
			_, err := gUniClient.Del(ctx, k).Result()
			if err != nil {
				return 0, err
			}
		}
	}

	return gUniClient.Del(ctx, key...).Result()
}

// SetNx Zero expiration means the key has no expiration time.
func SetNx(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool /*key is not exist*/, error) {
	return gUniClient.SetNX(ctx, key, value, expiration).Result()
}

func SetBytes(ctx context.Context, key string, value []byte, expiration time.Duration) (string, error) {
	return Set(ctx, key, value, expiration)
}

func LPush(ctx context.Context, key string, values ...interface{}) (int64, error) {
	return gUniClient.LPush(ctx, key, values...).Result()
}

func LRem(ctx context.Context, key string, count int64, value string) (int64, error) {
	return gUniClient.LRem(ctx, key, count, value).Result()
}

func LPop(ctx context.Context, key string) (string, error) {
	return gUniClient.LPop(ctx, key).Result()
}

func LRange(ctx context.Context, key string, begin, end int64) ([]string, error) {
	return gUniClient.LRange(ctx, key, begin, end).Result()
}

func ScriptLoad(ctx context.Context, script string) (string, error) {
	return gUniClient.ScriptLoad(ctx, script).Result()
}

func EvalSha(ctx context.Context, sha1 string, keys []string, args ...interface{}) (interface{}, error) {
	return gUniClient.EvalSha(ctx, sha1, keys, args...).Result()
}

func EvalShaInt64(ctx context.Context, sha1 string, keys []string, args ...interface{}) (int64, error) {
	return gUniClient.EvalSha(ctx, sha1, keys, args...).Int64()
}

func Eval(ctx context.Context, script string, keys []string, args ...interface{}) (interface{}, error) {
	return gUniClient.Eval(ctx, script, keys, args...).Result()
}

// Expire returns bool (0/1) ,false: the key not exist
func Expire(ctx context.Context, key string, t time.Duration) (bool, error) {
	return gUniClient.Expire(ctx, key, t).Result()
}

// Incr return result
func Incr(ctx context.Context, key string) (int64, error) {
	return gUniClient.Incr(ctx, key).Result()
}

const INCR_WITH_EXPIRATION_SCRIPT = `
local result = redis.call("incr", KEYS[1])
if tonumber(result) == 1 then
	result = redis.call("expire", KEYS[1], ARGV[1])
end
return result
`

func IncrWithExpiration(ctx context.Context, key string, expiration time.Duration) (int64, error) {
	return gUniClient.Eval(ctx, INCR_WITH_EXPIRATION_SCRIPT, []string{key}, expiration.Seconds()).Int64()
}

func Publish(ctx context.Context, channel string, value string) (int64, error) {
	return gUniClient.Publish(ctx, channel, value).Result()
}

// Subscribe 的返回具体使用看其源码示例
func Subscribe(ctx context.Context, channels ...string) *goRedis.PubSub {
	return gUniClient.Subscribe(ctx, channels...)
}

// Lock 返回true代表key不存在，加锁成功
// Lock 返回false代表key存在，加锁失败
func Lock(ctx context.Context, key string, expiration time.Duration) (bool, error) {
	return gUniClient.SetNX(ctx, key, "1", expiration).Result()
}

func UnLock(ctx context.Context, key string) (int64, error) {
	return gUniClient.Del(ctx, key).Result()
}

// LockV2 返回true代表key不存在，加锁成功
// LockV2 返回false代表key存在，加锁失败
func LockV2(ctx context.Context, key string, expiration time.Duration) (bool, string, error) {
	val := util.NewSnowflakeID()
	b, err := gUniClient.SetNX(ctx, key, val, expiration).Result()
	return b, val, err
}

func UnLockV2(ctx context.Context, key, val string) (int64, error) {
	v, err := gUniClient.Get(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	if v != val {
		return 0, errors.New("val not match, unlock error")
	}
	return gUniClient.Del(ctx, key).Result()
}

func Get(ctx context.Context, key string) (string, error) {
	return get(ctx, key).Result()
}

func GetBytes(ctx context.Context, key string) ([]byte, error) {
	return get(ctx, key).Bytes()
}

func GetString(ctx context.Context, key string) (string, error) {
	return get(ctx, key).Result()
}

func TTL(ctx context.Context, key string) (time.Duration, error) {
	return gUniClient.TTL(ctx, key).Result()
}

func Keys(ctx context.Context, pattern string) ([]string, error) {
	cfg := fconfig.DefaultConfig

	// 非集群模式下，直接scan
	if cfg.RedisMode != fconfig.REDIS_MODE_CLUSTER {
		var cursor uint64
		allKeys := make([]string, 0, 2)
		for {
			var keys []string
			var err error
			keys, cursor, err = gUniClient.Scan(ctx, cursor, pattern, 1000).Result()
			if err != nil {
				return nil, err
			}
			allKeys = append(allKeys, keys...)
			if cursor == 0 {
				break
			}
		}

		return allKeys, nil
	}

	// 集群模式下，需要逐个 node scan
	allKeys := make([]string, 0, 2)
	var mux sync.Mutex
	clusterClient := gUniClient.(*goRedis.ClusterClient)
	err := clusterClient.ForEachMaster(ctx, func(ctx context.Context, client *goRedis.Client) error {
		var cursor uint64
		for {
			var keys []string
			var err error
			keys, cursor, err = client.Scan(ctx, cursor, pattern, 1000).Result()
			if err != nil {
				return err
			}
			mux.Lock()
			allKeys = append(allKeys, keys...)
			mux.Unlock()
			if cursor == 0 {
				break
			}
		}
		return nil
	})

	return allKeys, err
}

func RedisNotFound(err error) bool {
	if err != nil {
		return err.Error() == "redis: nil"
	}

	return errors.Is(err, goRedis.Nil)
}

// GetOrSet 从缓存中获取数据，如果不存在则从fetcher中获取数据并设置到缓存中
func GetOrSet(ctx context.Context, cacheKey string, expiration time.Duration, fetcher func(ctx context.Context) ([]byte, *ferrors.SvrRspInfo)) ([]byte, *ferrors.SvrRspInfo) {
	rspInfo := ferrors.Ok()

	val, err := GetBytes(ctx, cacheKey)
	if err == nil && len(val) > 0 {
		return val, rspInfo
	}

	valRaw, rspInfo := fetcher(ctx)
	if !rspInfo.Valid() {
		return nil, rspInfo
	}

	_, err = SetBytes(ctx, cacheKey, valRaw, expiration)
	if err != nil {
		log.Errorc(ctx, "fredis GetOrSet Set failed: %s", err)
	}

	return valRaw, rspInfo
}

// GetOrSetCondition 如果condition为true, 则优先从缓存中获取数据，如果不存在则从fetcher中获取数据并设置到缓存中
func GetOrSetCondition(ctx context.Context, cacheKey string, expiration time.Duration, condition func() bool, fetcher func(ctx context.Context) ([]byte, *ferrors.SvrRspInfo)) ([]byte, *ferrors.SvrRspInfo) {
	useCache := condition()
	if !useCache {
		valRaw, rspInfo := fetcher(ctx)
		return valRaw, rspInfo
	}

	return GetOrSet(ctx, cacheKey, expiration, fetcher)
}

func ScanDel(ctx context.Context, keyPrefix string) error {
	allKeys, err := Keys(ctx, keyPrefix)
	if err != nil {
		return err
	}

	if len(allKeys) == 0 {
		return nil
	}

	// 集群模式下，需要逐个删除
	cfg := fconfig.DefaultConfig
	if cfg.RedisMode == fconfig.REDIS_MODE_CLUSTER {
		for _, key := range allKeys {
			_, err := gUniClient.Del(ctx, key).Result()
			if err != nil {
				return err
			}
		}
		return nil
	}

	_, err = gUniClient.Del(ctx, allKeys...).Result()
	return err
}

func get(ctx context.Context, key string) *goRedis.StringCmd {
	return gUniClient.Get(ctx, key)
}

func SIsMember(ctx context.Context, key string, member interface{}) (bool, error) {
	return gUniClient.SIsMember(ctx, key, member).Result()
}

func SCard(ctx context.Context, key string) (int64, error) {
	return gUniClient.SCard(ctx, key).Result()
}

func SAdd(ctx context.Context, key string, member interface{}) (int64, error) {
	return gUniClient.SAdd(ctx, key, member).Result()
}

func SRem(ctx context.Context, key string, members ...interface{}) (int64, error) {
	return gUniClient.SRem(ctx, key, members).Result()
}

func BRPop(ctx context.Context, wait time.Duration, queue string) ([]string, error) {
	return gUniClient.BRPop(ctx, wait, queue).Result()
}

func PFAdd(ctx context.Context, key string, els ...interface{}) (int64, error) {
	return gUniClient.PFAdd(ctx, key, els).Result()
}

func PFCount(ctx context.Context, keys ...string) (int64, error) {
	return gUniClient.PFCount(ctx, keys...).Result()
}

func HGetAll(ctx context.Context, key string) (map[string]string, error) {
	return gUniClient.HGetAll(ctx, key).Result()
}
func HIncrBy(ctx context.Context, key, field string, incr int64) (int64, error) {
	return gUniClient.HIncrBy(ctx, key, field, incr).Result()
}
func HDel(ctx context.Context, key string, fields ...string) (int64, error) {
	return gUniClient.HDel(ctx, key, fields...).Result()
}

func flushAll(ctx context.Context) error {
	return gUniClient.FlushAll(ctx).Err()
}

func Append(ctx context.Context, key, value string) error {
	return gUniClient.Append(ctx, key, value).Err()
}

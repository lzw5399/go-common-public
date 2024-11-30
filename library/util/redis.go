package util

import (
	"errors"

	goRedis "github.com/go-redis/redis/v8"
)

func RedisNotFound(err error) bool {
	return errors.Is(err, goRedis.Nil)
}

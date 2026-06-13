package redis

import (
	"context"
	"os"
	goredis "github.com/redis/go-redis/v9"
)

var Ctx = context.Background()

var Client = goredis.NewClient(&goredis.Options{
	Addr: os.Getenv("REDIS_ADDR"),
})
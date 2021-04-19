package database

import (
	"github.com/go-redis/redis/v8"
)

var client *redis.Client

func initRedis() {
	// initializing redis
	opt, err := redis.ParseURL("redis://<user>:<pass>@localhost:6379/<db>")
	if err != nil {
		panic(err)
	}

	client = redis.NewClient(opt)
}

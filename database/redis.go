package database

import (
	"context"
	"github.com/go-redis/redis/v8"
)

var (
	Client = &redisClient{}
	Ctx    = context.TODO()
)

type redisClient struct {
	c *redis.Client
}

func InitRedis() *redisClient {
	// initializing redis
	c := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	_, err := c.Ping(Ctx).Result()
	if err != nil {
		panic(err)
	}

	Client.c = c
	return Client
}

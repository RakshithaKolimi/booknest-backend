package controller

import "github.com/go-redis/redis"

var redisClient *redis.Client

func SetRedisClient(client *redis.Client) {
	redisClient = client
}

func getRedisClient() *redis.Client {
	return redisClient
}

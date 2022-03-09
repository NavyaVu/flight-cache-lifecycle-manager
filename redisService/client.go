package redisService

import (
	redisV8 "github.com/go-redis/redis/v8"
	"golang.org/x/net/context"
)

var RedisClientType RedisClient

type RedisClient interface {
	LifeCycleManager(keysToBeDeleted []string, client *redisV8.Client) string
	GetAllKeys(client *redisV8.Client) []string
}

type RealRedisClient struct {
}

func (redisClient RealRedisClient) GetAllKeys(client *redisV8.Client) []string {
	k := client.Keys(context.Background(), "*")
	return k.Val()
}
func (redisClient RealRedisClient) LifeCycleManager(keysToBeDeleted []string, client *redisV8.Client) string {

	if len(keysToBeDeleted) >= 1 {
		client.Del(context.Background(), keysToBeDeleted...)
		return "Keys deleted"
	}

	return "Nothing is deleted as all the keys have departure date after today's date or the Redis is empty"

}

package redisService

import (
	"flight-cache-lifecycle-manager/models"
	"fmt"
	redisV8 "github.com/go-redis/redis/v8"
	"golang.org/x/net/context"
	"log"
	"time"
)

var RedisClientType RedisClient

type RedisClient interface {
	LifeCycleManager(keysToBeDeleted []string, client *redisV8.Client) string
	GetAllKeys(client *redisV8.Client) []string
	Query(CacheEntryKey string, client *redisV8.Client) (*models.CacheEntry, error)
	AddEntry(CacheEntry *models.CacheEntry, client *redisV8.Client, ttlInMinutes int) error
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

func (redisClient RealRedisClient) Query(CacheEntryKey string, client *redisV8.Client) (*models.CacheEntry, error) {
	val, err := client.Get(context.Background(), CacheEntryKey).Result()
	if err != nil {
		fmt.Println(err)
	}

	return &models.CacheEntry{
		Key:   CacheEntryKey,
		Value: val,
	}, err

}

func (redisClient RealRedisClient) AddEntry(CacheEntry *models.CacheEntry, client *redisV8.Client, ttlInMinutes int) error {
	err := client.Set(context.Background(), CacheEntry.Key, CacheEntry.Value, time.Duration(int(time.Minute)*ttlInMinutes)).Err()
	if err != nil {
		log.Fatalln(err.Error())
	}
	log.Println("Entry added with key : ", CacheEntry.Key, " at ", time.Now(), "with ttl as ", ttlInMinutes,
		" minutes")
	return err
}

package redisService

import (
	"fmt"
	redisV8 "github.com/go-redis/redis/v8"
	"golang.org/x/net/context"
	"log"
	"regexp"
	"strings"
	"time"
)

var RedisClientType RedisClient

type RedisClient interface {
	LifeCycleManager(client *redisV8.Client) (string, error)
}

type RealRedisClient struct {
}

func (redisClient RealRedisClient) LifeCycleManager(client *redisV8.Client) (string, error) {

	keys := client.Keys(context.Background(), "*")

	var (
		keysToBeDeleted []string
		err             error
	)

	for _, j := range keys.Val() {
		dt := time.Now()
		//y, m, d := dt.Date()
		//dtf := fmt.Sprintf("%d-%d-%d", y,int(m),d )
		//fmt.Println(dtf)

		re := regexp.MustCompile("[0-9]+")
		reStringArray := re.FindAllString(j, 3)
		if len(reStringArray) < 3 {
			fmt.Println("deleted", "key:", j)
			keysToBeDeleted = append(keysToBeDeleted, j)
		} else {

			keySt := strings.Join(reStringArray[:], "-")
			x, err := convertDate(keySt)

			if err != nil {
				log.Println(err)
			}

			compare := dt.Before(x)
			if compare {
				fmt.Println("Date is after today's date, so no deletion for key: ", j)
			} else {
				fmt.Println("deleted", "key:", j)
				keysToBeDeleted = append(keysToBeDeleted, j)
			}
		}

	}

	if len(keysToBeDeleted) >= 1 {
		client.Del(context.Background(), keysToBeDeleted...)
		return "Keys deleted", nil
		//client.Keys(context.Background(), "*")
	}

	return "Nothing is deleted as all the keys have departure date after today's date or the Redis is empty", err

}

func convertDate(date string) (time.Time, error) {
	layout := "2006-01-02"
	t, err := time.Parse(layout, date)

	if err != nil {
		log.Println(err)
	}

	return t, err
}

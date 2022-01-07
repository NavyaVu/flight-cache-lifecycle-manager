package main

import (
	"encoding/json"
	"flight-cache-lifecycle-manager/config"
	"flight-cache-lifecycle-manager/redisService"
	"github.com/aws/aws-lambda-go/lambda"
	redisV8 "github.com/go-redis/redis/v8"
	"github.com/magiconair/properties"
	"log"
	"strings"
)

var (
	redisClient           *redisV8.Client
	flightCacheProperties *properties.Properties
)

func init() {
	flightCacheProperties = config.LoadProperties()
	redisClient = getRedisClient(flightCacheProperties)
	redisService.RedisClientType = redisService.RealRedisClient{}
}

func main() {
	lambda.Start(HandleRequest)
}

func HandleRequest(input interface{}) (interface{}, error) {

	var err error
	inputRequest, err := json.Marshal(input)
	if err != nil {
		log.Println("unable to marshal input to json")
	}

	inputRequestAsString := string(inputRequest)
	containsHeader := strings.Contains(inputRequestAsString, "headers")
	log.Println("does the request contain header ", containsHeader)
	if !containsHeader {
		log.Println("calling lambda for request ", inputRequestAsString)

		redisClientType := redisService.RedisClientType
		response, err := redisClientType.LifeCycleManager(redisClient)
		if err != nil {
			return err.Error(), err
		} else {
			log.Println("Response Body", response)
			return response, err
		}
	}
	return "Nothing Executed", err
}

func getRedisClient(p *properties.Properties) *redisV8.Client {
	redisAddr, _ := p.Get("redis-addr-port-AWS")
	return redisV8.NewClient(&redisV8.Options{
		Addr:     redisAddr,
		Password: "",
		DB:       0,
	})
}

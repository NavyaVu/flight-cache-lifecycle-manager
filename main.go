package main

import (
	"crypto/tls"
	"encoding/json"
	"flight-cache-lifecycle-manager/config"
	"flight-cache-lifecycle-manager/openSearch"
	"flight-cache-lifecycle-manager/redisService"
	"flight-cache-lifecycle-manager/service"
	redisV8 "github.com/go-redis/redis/v8"
	"github.com/magiconair/properties"
	"github.com/opensearch-project/opensearch-go"
	"log"
	"net/http"
	"strings"
)

var (
	redisClient                  *redisV8.Client
	flightCacheManagerProperties *properties.Properties
	openSearchClient             *opensearch.Client
	cacheManager                 service.CacheManagerImpl
	dbManager                    service.DbManagerImpl
)

func init() {
	flightCacheManagerProperties = config.LoadProperties()
	redisClient = getRedisClient(flightCacheManagerProperties)
	openSearchClient, _ = getOpenSearchClient(flightCacheManagerProperties)
	openSearch.OpenSearchClientType = openSearch.RealOpenSearchClient{}
	redisService.RedisClientType = redisService.RealRedisClient{}

	cacheManager = service.CacheManagerImpl{
		Client:     redisClient,
		ClientType: redisService.RedisClientType,
	}

	dbManager = service.DbManagerImpl{
		OSClient:     openSearchClient,
		OSClientType: openSearch.OpenSearchClientType,
	}
}

func main() {

	var (
		resFormCache      = make(chan string)
		resErrorFormCache = make(chan error)
		resFormDb         = make(chan string)
		resErrorFormDb    = make(chan error)
	)
	go dbManager.ManageEntries(flightCacheManagerProperties, resFormCache, resErrorFormCache)
	if resErrorFormCache != nil {
		log.Println(resErrorFormCache)
	}
	go cacheManager.ManageEntries(flightCacheManagerProperties, resFormDb, resErrorFormDb)
	if resErrorFormDb != nil {
		log.Println(resErrorFormDb)
	}
	log.Println(<-resFormCache)
	log.Println(<-resFormDb)

	//lambda.Start(HandleRequest)
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

		var (
			resFormCache      = make(chan string)
			resErrorFormCache = make(chan error)
			resFormDb         = make(chan string)
			resErrorFormDb    = make(chan error)
		)
		go dbManager.ManageEntries(flightCacheManagerProperties, resFormCache, resErrorFormCache)
		if resErrorFormCache != nil {
			log.Println(resErrorFormCache)
		}
		go cacheManager.ManageEntries(flightCacheManagerProperties, resFormDb, resErrorFormDb)
		if resErrorFormDb != nil {
			log.Println(resErrorFormDb)
		}
		log.Println(<-resFormCache)
		log.Println(<-resFormDb)

	}
	return "Nothing Executed", err
}

func getRedisClient(p *properties.Properties) *redisV8.Client {
	redisAddr, _ := p.Get("redis-addr-port-AWS")
	redisAddr = "localhost:6379"
	return redisV8.NewClient(&redisV8.Options{
		Addr:     redisAddr,
		Password: "",
		DB:       0,
	})
}

func getOpenSearchClient(p *properties.Properties) (*opensearch.Client, error) {

	openSearchURL, _ := p.Get("openSearch-endpoint-URL")
	openSearchUsername, _ := p.Get("openSearch-Username")
	openSearchPassword, _ := p.Get("openSearch-Password")
	client, err := opensearch.NewClient(opensearch.Config{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		Addresses: []string{openSearchURL},
		Username:  openSearchUsername,
		Password:  openSearchPassword,
	})

	return client, err
}

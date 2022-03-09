package service

import (
	"flight-cache-lifecycle-manager/models"
	"flight-cache-lifecycle-manager/openSearch"
	"flight-cache-lifecycle-manager/redisService"
	"fmt"
	redisV8 "github.com/go-redis/redis/v8"
	"github.com/magiconair/properties"
	"github.com/opensearch-project/opensearch-go"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type ManagerService interface {
	ManageEntries(*properties.Properties) ([]string, error)
}

type DbManagerImpl struct {
	OSClient     *opensearch.Client
	OSClientType openSearch.OpenSearchClient
}

type CacheManagerImpl struct {
	Client     *redisV8.Client
	ClientType redisService.RedisClient
}

func checkIfKeyHasExpired(key *string, timestamp time.Time, prop *properties.Properties) (bool, error) {
	var (
		isKeyExpired bool
	)
	splitsKey := strings.Split(*key, "-")
	if len(splitsKey) > 3 {
		airlineCodeWithSource := fmt.Sprintf(splitsKey[2] + "-" + splitsKey[3])
		ttlValue, found := prop.Get(models.AIRLINE_EXP + "-" + airlineCodeWithSource)
		if found {
			intConv, err := strconv.Atoi(ttlValue)

			if err != nil {
				log.Println(err)
			}

			ttl := time.Duration(intConv) * time.Minute

			isKeyExpired = checkForKeysDeletion(timestamp, ttl)
		} else {
			log.Println("No config found for this key:", *key)
			log.Println("Checking for the deletion of keys with past departure dates")
			isKeyExpired = checkIfKeyHasPastDeptDate(*key)
		}
	}

	return isKeyExpired, nil
}

func (db *DbManagerImpl) ManageEntries(prop *properties.Properties) ([]string, error) {
	var (
		keysTobeDeleted []string
	)

	m, err := db.OSClientType.GetAllKeysFromOS(db.OSClient)

	if err != nil {
		log.Println(err)
	} else {
		for key, timestamp := range m {
			if len(key) < models.MIN_KEY_LENGTH {
				log.Println("Invalid Key to be deleted: ", key)
				keysTobeDeleted = append(keysTobeDeleted, key)
			} else {
				isKeyExpired, err := checkIfKeyHasExpired(&key, timestamp, prop)
				if isKeyExpired && err == nil {
					log.Println("Key to be deleted: ", key)
					keysTobeDeleted = append(keysTobeDeleted, key)
				}
			}
		}
	}

	if len(keysTobeDeleted) > 0 {
		statusCode, err := db.OSClientType.DeleteEntry(keysTobeDeleted, db.OSClient)
		if err != nil {
			log.Println(err)
		}
		log.Println(statusCode)
	}

	return keysTobeDeleted, err
}

func checkForKeysDeletion(t time.Time, y time.Duration) bool {
	x := t.Add(y)
	timeNow := time.Now().UTC()

	b := x.Before(timeNow)
	fmt.Println(b)

	return b
}

func (cache *CacheManagerImpl) ManageEntries(prop *properties.Properties) ([]string, error) {
	keys := cache.ClientType.GetAllKeys(cache.Client)

	keysToBeDeleted, err := deletionBasedOnDepDate(keys)
	r := cache.ClientType.LifeCycleManager(keysToBeDeleted, cache.Client)
	fmt.Println(r)
	return keysToBeDeleted, err
}

func checkIfKeyHasPastDeptDate(key string) bool {

	isPastDeptDate := false

	dt := time.Now().UTC()

	re := regexp.MustCompile("[0-9]+")
	reStringArray := re.FindAllString(key, 3)
	if len(reStringArray) < 3 {
		log.Println("deleted", "key:", key)
		isPastDeptDate = true
	} else {

		keySt := strings.Join(reStringArray[:], "-")
		x, err := convertDate(keySt)

		if err != nil {
			log.Println(err)
		} else {

			check := dt.Before(x)
			if check {
				log.Println("Date is after today's date, so no deletion for key: ", key)
			} else {
				log.Println("deleted", "key:", key)
				isPastDeptDate = true
			}
		}
	}
	return isPastDeptDate
}

func deletionBasedOnDepDate(keys []string) ([]string, error) {

	fmt.Println("Deletion of Keys based on Departure Date")

	var (
		keysToBeDeleted []string
		err             error
	)

	for _, j := range keys {
		dt := time.Now().UTC()

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
			} else {

				compare := dt.Before(x)
				if compare {
					fmt.Println("Date is after today's date, so no deletion for key: ", j)
				} else {
					fmt.Println("deleted", "key:", j)
					keysToBeDeleted = append(keysToBeDeleted, j)
				}
			}
		}

	}

	return keysToBeDeleted, err
}

func convertDate(date string) (time.Time, error) {
	layout := "2006-01-02"
	t, err := time.Parse(layout, date)

	if err != nil {
		log.Println(err)
	}

	return t, err
}

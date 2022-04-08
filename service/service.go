package service

import (
	"encoding/json"
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
	ManageEntries(*properties.Properties, chan<- string, chan<- error)
}

type DbManagerImpl struct {
	OSClient     *opensearch.Client
	OSClientType openSearch.OpenSearchClient
}

type CacheManagerImpl struct {
	Client     *redisV8.Client
	ClientType redisService.RedisClient
}

func checkIfKeyIsValidBasedOnConfig(key *string, timestamp time.Time, prop *properties.Properties) (bool, error) {
	var (
		isKeyValid bool
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

			isKeyValid = checkForValidKeyBasedOnCurrentTime(timestamp, ttl)
		}
	}

	return isKeyValid, nil
}

func checkIfKeyIsValidBasedOnOfferValidity(id *string, combinations *[]models.Combination, prop *properties.Properties, dbType string) ([]string, error) {

	//updateStringForValidField := "ctx._source.combinations[2].additionalParams.isValid= params.valid"
	var updateString []string
	var err error

	for i, combination := range *combinations {
		value := combination.AdditionalParams["offerValidity"]
		if len(value) != 0 {
			fmt.Println("found the offer with offerValidity")
			valueTimeConverted, err := convertDate(value)
			if err != nil {
				log.Println(err.Error())
			} else {

				check := checkForValidKeyBasedOnCurrentTime(valueTimeConverted, 0)

				//for redis
				if dbType == "redis" {
					validityUpdateForCacheBasedOnKey(check, &combination)
				} else if dbType == "openSearch" {
					updateString = append(updateString, stringAppenderForValidity(i, check))
				} else {
					log.Println("None of the db matches ")
				}

			}
		} else {
			fmt.Println("found the offer without offerValidity")
			//update time stamp
			var timeStamp string
			timeStamp = combination.AdditionalParams["updatedTimeStamp"]
			if len(timeStamp) == 0 {
				timeStamp = combination.AdditionalParams["insertionTimeStamp"]
			}
			ts, err := convertDate(timeStamp)

			if len(timeStamp) != 0 {
				if err != nil {
					log.Println(err.Error())
				} else {
					check, err := checkIfKeyIsValidBasedOnConfig(id, ts, prop)
					if err != nil {
						log.Println(err.Error())
					} else {
						if dbType == "redis" {
							validityUpdateForCacheBasedOnKey(check, &combination)
						} else if dbType == "openSearch" {
							updateString = append(updateString, stringAppenderForValidity(i, check))
						} else {
							log.Println("None of the db matches ")
						}
					}
				}
			}
		}
	}

	//call update in es with id,string
	return updateString, err
}

func stringAppenderForValidity(i int, check bool) string {
	if check {
		return models.OpenSearchUpdateStringCombinations + strconv.Itoa(i) + models.OpenSearchUpdateStringAdditionalParam + "valid"
	} else {
		return models.OpenSearchUpdateStringCombinations + strconv.Itoa(i) + models.OpenSearchUpdateStringAdditionalParam + "notValid"
	}
}

func (db *DbManagerImpl) ManageEntries(prop *properties.Properties, response chan<- string, resError chan<- error) {
	var (
		keysTobeDeleted []string
	)

	osResponse, err := db.OSClientType.GetResponseFromOS(db.OSClient)

	if err != nil {
		log.Println(err)
	} else {

		for _, hit := range osResponse.Hits.Hits {

			if len(hit.ID) < models.MIN_KEY_LENGTH {
				log.Println("Invalid Key to be deleted: ", hit.ID)
				keysTobeDeleted = append(keysTobeDeleted, hit.ID)
			} else if checkIfKeyHasPastDeptDate(hit.ID) {
				keysTobeDeleted = append(keysTobeDeleted, hit.ID)
			} else {

				str, err := checkIfKeyIsValidBasedOnOfferValidity(&hit.ID, &hit.Source.Combinations, prop, "openSearch")

				if err != nil {
					log.Println(err.Error())
				} else {
					if len(str) > 0 {
						updateString := strings.Join(str, ";")
						db.OSClientType.UpdateValidityBasedOnOfferValidity(db.OSClient, hit.ID, updateString)
					}
				}

			}

		}
	}
	if len(keysTobeDeleted) > 0 {
		statusCode, err := db.OSClientType.DeleteEntry(keysTobeDeleted, db.OSClient)
		if err != nil {
			log.Println(err)
		}
		log.Println(statusCode, ": Status code for deleted entries")
	}

	response <- "Managed entries in Database"
	resError <- err
}

func validityUpdateForCacheBasedOnKey(check bool, com *models.Combination) {

	if check {
		com.AdditionalParams["isValid"] = "true"
	} else {
		com.AdditionalParams["isValid"] = "false"
	}
}

func checkForValidKeyBasedOnCurrentTime(t time.Time, y time.Duration) bool {
	x := t.Add(y)
	timeNow := time.Now().UTC()

	b := x.After(timeNow)
	fmt.Println(b)

	return b
}

func (cache *CacheManagerImpl) ManageEntries(prop *properties.Properties, response chan<- string, resError chan<- error) {

	var keysToBeDeleted []string
	var err error
	//var fResult  *models.Result
	var ce *models.CacheEntry
	keys := cache.ClientType.GetAllKeys(cache.Client)

	for _, key := range keys {
		//var fResult  *models.Result
		if checkIfKeyHasPastDeptDate(key) {
			keysToBeDeleted = append(keysToBeDeleted, key)
		} else {
			cacheEntry, err := cache.ClientType.Query(key, cache.Client)

			if err != nil {
				log.Println(err.Error())
			} else {
				res, err := loadResultFromCache(cacheEntry.Value)

				if err != nil {
					log.Println(err.Error())
				} else {
					_, err := checkIfKeyIsValidBasedOnOfferValidity(&key, &res.Combinations, prop, "redis")

					var finalRes *models.Result

					finalRes = &models.Result{
						Routes:           res.Routes,
						Segments:         res.Segments,
						Combinations:     res.Combinations,
						Ancillaries:      res.Ancillaries,
						AdditionalParams: res.AdditionalParams,
					}
					resInBytes, err := json.Marshal(&finalRes)

					ce = &models.CacheEntry{
						Key:   cacheEntry.Key,
						Value: string(resInBytes),
					}

					err = cache.ClientType.AddEntry(ce, cache.Client, 0)

					if err != nil {
						log.Println(err.Error())
					}
				}

			}

		}

		//call redis to update the key

	}

	r := cache.ClientType.LifeCycleManager(keysToBeDeleted, cache.Client)
	fmt.Println(r)
	response <- "Managed entries in Cache"
	resError <- err
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

func convertDate(d string) (time.Time, error) {

	layout := "2006-01-02"
	if strings.Contains(d, "T") {
		layout = "2006-01-02T15:04:05.000Z"
	}
	t, err := time.Parse(layout, d)

	if err != nil {
		log.Println(err)
	}

	return t, err
}

func loadResultFromCache(cacheValue string) (*models.Result, error) {
	var result *models.Result
	err := json.Unmarshal([]byte(cacheValue), &result)
	if err != nil {
		log.Println("Unable to unmarshal string response to tfm response ", err.Error())
	}

	return result, err

}

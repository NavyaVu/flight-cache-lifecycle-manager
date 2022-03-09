package service

import (
	redisV8 "github.com/go-redis/redis/v8"
	"github.com/magiconair/properties"
	"github.com/magiconair/properties/assert"
	"github.com/opensearch-project/opensearch-go"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type MockOpenSearchClient struct {
}

func (osMock MockOpenSearchClient) DeleteEntry(cacheEntryKeys []string, client *opensearch.Client) (int, error) {

	return 200, nil
}

func (osMock MockOpenSearchClient) GetAllKeysFromOS(*opensearch.Client) (map[string]time.Time, error) {

	var k = make(map[string]time.Time)
	y := time.Now()
	k["AMS-NYC-EK-NDC-2022-06-04-O"] = y.AddDate(0, -3, -3)
	k["AMS-NYC-EK-NDC-2022-06-03-O"] = y.AddDate(0, -3, -2)
	k["AMS-NYC-AF-OTH-2022-06-02-O"] = y.AddDate(0, -3, -3)
	k["AMS-NYC"] = y.AddDate(0, -3, -3)
	k["AMS-NYC-AF-OTH-2022-06-04-O"] = y.AddDate(0, 3, 3)
	return k, nil
}

type MockRedisClient struct {
}

func (mCache MockRedisClient) LifeCycleManager(keysToBeDeleted []string, client *redisV8.Client) string {
	return "Mock redis client"
}

func (mCache MockRedisClient) GetAllKeys(client *redisV8.Client) []string {
	keys := []string{"AMS-NYC-EK-NDC-2022-01-04-O", "AMS-NYC-EK-NDC-2022-02-04-O", "AMS-NYC", "AMS-NYC-EK-NDC-2022-06-04-O"}
	return keys
}

func loadProperties() *properties.Properties {
	myDir, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}
	log.Println(myDir)

	propFile, err := filepath.Abs("../resources/service.properties")
	if err != nil {
		log.Panicln("PropFile file not found at ", propFile, " error: ", err.Error())
	}

	propertiesInstance := properties.MustLoadFile(propFile, properties.UTF8)

	return propertiesInstance
}

func Test_ManageEntriesCache(t *testing.T) {
	cacheClientType := MockRedisClient{}
	cacheManagerService := CacheManagerImpl{
		Client:     nil,
		ClientType: cacheClientType,
	}
	b, err := cacheManagerService.ManageEntries(nil)

	assert.Equal(t, 3, len(b))
	assert.Equal(t, nil, err)

}

func Test_ManageEntries(t *testing.T) {

	//case-1 : no config found but got inserted long back ?

	prop := loadProperties()
	//prepare or setup
	dbClientType := MockOpenSearchClient{}
	dbMangerService := DbManagerImpl{
		OSClient:     nil,
		OSClientType: dbClientType,
	}

	//execute
	keysDeleted, err := dbMangerService.ManageEntries(prop)
	log.Println("keys to be deleted:", keysDeleted)
	//validations
	assert.Equal(t, len(keysDeleted), 2)
	assert.Equal(t, err, nil)
}

func Test_deletionBasedOnDepDate(t *testing.T) {

	keys := []string{"AMS-NYC-EK-NDC-2022-06-04-O", "AMS-NYC-EK-NDC-2022-06-03-O", "AMS-NYC-EK-NDC-2022-06-02-O"}
	k, _ := deletionBasedOnDepDate(keys)

	assert.Equal(t, 0, len(k))

}

func Test_deletionBasedOnDepDateWithError(t *testing.T) {

	keys := []string{"AMS-NYC"}
	k, _ := deletionBasedOnDepDate(keys)

	assert.Equal(t, 1, len(k))

}

func Test_deletionBasedOnDepDateWithoutError(t *testing.T) {

	keys := []string{"AMS-NYC-EK-NDC-2022-02-02-O", "AMS-NYC-EK-NDC-2022-02-03-O", "AMS-NYC-EK-NDC-2022-02-04-O"}
	k, _ := deletionBasedOnDepDate(keys)

	assert.Equal(t, 3, len(k))

}

func Test_checkForKeysDeletionTrue(t *testing.T) {

	y := time.Now()
	x := y.AddDate(0, -3, -3)

	z := time.Duration(6) * time.Minute

	b := checkForKeysDeletion(x, z)

	assert.Equal(t, true, b)

}

func Test_checkForKeysDeletionFalse(t *testing.T) {

	y := time.Now()
	x := y.AddDate(0, 3, 3)

	z := time.Duration(6) * time.Minute

	b := checkForKeysDeletion(x, z)

	assert.Equal(t, false, b)

}

func Test_checkIfKeyHasPastDeptDate(t *testing.T) {
	b := checkIfKeyHasPastDeptDate("AMS-NYC-EK-NDC-2022-02-02-O")

	assert.Equal(t, true, b)
}

func Test_checkIfKeyHasPastDeptDateLen(t *testing.T) {
	b := checkIfKeyHasPastDeptDate("AMS-NYC")

	assert.Equal(t, true, b)
}

func Test_checkIfKeyHasPastDeptDateFalse(t *testing.T) {
	b := checkIfKeyHasPastDeptDate("AMS-NYC-EK-NDC-2022-06-02-O")

	assert.Equal(t, false, b)
}

func Test_checkIfKeyHasExpired(t *testing.T) {

	prop := loadProperties()

	y := time.Now()
	x := y.AddDate(0, -3, -3)
	k := "AMS-NYC-AF-OTH-2022-06-02-O"

	b, err := checkIfKeyHasExpired(&k, x, prop)

	assert.Equal(t, true, b)
	assert.Equal(t, nil, err)
}

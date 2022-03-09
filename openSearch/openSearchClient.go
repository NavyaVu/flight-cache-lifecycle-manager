package openSearch

import (
	"bytes"
	"context"
	"encoding/json"
	"flight-cache-lifecycle-manager/models"
	"fmt"
	"github.com/opensearch-project/opensearch-go"
	"github.com/opensearch-project/opensearch-go/opensearchapi"
	"log"
	"os"
	"strings"
	"time"
)

var OpenSearchClientType OpenSearchClient

type OpenSearchClient interface {
	DeleteEntry([]string, *opensearch.Client) (int, error)
	GetAllKeysFromOS(*opensearch.Client) (map[string]time.Time, error)
}

type RealOpenSearchClient struct {
}

func (osClient RealOpenSearchClient) DeleteEntry(keys []string, client *opensearch.Client) (int, error) {

	var (
		buf       bytes.Buffer
		indexName []string
		val       = keys
	)
	//val := []string{"AMS-NYC-KL-NDC-2021-11-23-O", "AMS-NYC-KL-NDC-2021-11-22-O"}

	for _, s := range val {
		log.Println("Values to be deleted in os", s)
	}

	query := map[string]interface{}{
		"query": map[string]interface{}{
			"ids": map[string]interface{}{
				"values": &val,
			},
		},
	}
	indexName = append(indexName, models.IndexName)

	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		log.Fatalf("Error encoding query: %s", err)
	}

	deleteReq := opensearchapi.DeleteByQueryRequest{
		Index:     indexName,
		Body:      &buf,
		Conflicts: "proceed",
	}

	res, err := deleteReq.Do(context.Background(), client)
	if err != nil {
		fmt.Println("failed to delete the document ", err)
		os.Exit(1)
	}
	fmt.Println(res)

	return res.StatusCode, nil
}

func (osClient RealOpenSearchClient) GetAllKeysFromOS(client *opensearch.Client) (map[string]time.Time, error) {

	//query for match all to get the id's

	var (
		indexName          []string
		openSearchResponse models.OpenSearchResponse
		kT                 = make(map[string]time.Time)
	)
	indexName = append(indexName, models.IndexName)

	bodyForSearchRequest := strings.NewReader(`{
		"query": {
			"match_all": {}
		}
	}`)

	sR := opensearchapi.SearchRequest{
		Index: indexName,
		Body:  bodyForSearchRequest,
	}

	res, err := sR.Do(context.Background(), client)

	if err != nil {
		log.Println("Error", err)
	}
	fmt.Println("Response during Search:", res.Status())

	if err := json.NewDecoder(res.Body).Decode(&openSearchResponse); err != nil {
		log.Fatalf("Error encoding query: %s", err)
	}

	for _, hit := range openSearchResponse.Hits.Hits {
		kT[hit.ID] = hit.Source.TimeStamp
	}

	return kT, err
}

package openSearch

import (
	"bytes"
	"context"
	"encoding/json"
	"flight-cache-lifecycle-manager/models"
	"github.com/opensearch-project/opensearch-go"
	"github.com/opensearch-project/opensearch-go/opensearchapi"
	"log"
	"os"
	"strings"
)

var OpenSearchClientType OpenSearchClient

type OpenSearchClient interface {
	DeleteEntry([]string, *opensearch.Client) (int, error)
	GetResponseFromOS(*opensearch.Client) (*models.OpenSearchResponse, error)
	UpdateValidityBasedOnOfferValidity(client *opensearch.Client, id string, body string) int
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
		log.Println("failed to delete the document ", err)
		os.Exit(1)
	}
	log.Println(res)

	return res.StatusCode, nil
}

func (osClient RealOpenSearchClient) GetResponseFromOS(client *opensearch.Client) (*models.OpenSearchResponse, error) {

	//query for match all to get the id's

	var (
		indexName          []string
		openSearchResponse models.OpenSearchResponse
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
	log.Println("Response during Search:", res.Status())

	if err := json.NewDecoder(res.Body).Decode(&openSearchResponse); err != nil {
		log.Fatalf("Error encoding query: %s", err)
	}

	return &openSearchResponse, err
}

func (osClient RealOpenSearchClient) UpdateValidityBasedOnOfferValidity(client *opensearch.Client, id string, body string) int {

	var buf bytes.Buffer

	script := map[string]interface{}{
		"script": map[string]interface{}{
			"source": body,
			"lang":   "painless",
			"params": map[string]string{
				"valid":    "true",
				"notValid": "false",
			},
		},
	}

	if err := json.NewEncoder(&buf).Encode(script); err != nil {
		log.Fatalf("Error encoding query: %s", err)
	}

	y := opensearchapi.UpdateRequest{
		Index:      models.IndexName,
		DocumentID: id,
		Body:       &buf,
	}

	res, err := y.Do(context.Background(), client)

	if err != nil {
		log.Println(err.Error())
	}
	log.Println(res.StatusCode, "Status code for Key: ", id)

	return res.StatusCode
}

package main

import (
	"context"
	"log"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/elastic/go-elasticsearch/v7/esapi"
)

var ESClient *elasticsearch.Client

const SearchIndex = "user"

func ESClientConnection() {
	client, err := elasticsearch.NewDefaultClient()
	if err != nil {
		log.Fatal(err)
	}

	ESClient = client
}

func ESCreateIndexIfNotExist() {
	_, err := esapi.IndicesExistsRequest{
		Index: []string{SearchIndex},
	}.Do(context.Background(), ESClient)

	if err != nil {
		ESClient.Indices.Create(SearchIndex)
	}
}

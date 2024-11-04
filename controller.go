package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/elastic/go-elasticsearch/v7/esapi"
)

type User struct {
	ID        string `json:"id,omitempty"`
	FirstName string `json:"first_name,omitempty"`
	LastName  string `json:"last_name,omitempty"`
	Email     string `json:"email,omitempty"`
}

func GetUser(w http.ResponseWriter, r *http.Request) {
	search := r.URL.Query().Get("search")

	var query map[string]interface{}
	if search != "" {
		query = map[string]interface{}{
			"query": map[string]interface{}{
				"multi_match": map[string]interface{}{
					"query":  search,
					"fields": []string{"first_name", "last_name", "email"},
				},
			},
		}
	} else {
		query = map[string]interface{}{
			"query": map[string]interface{}{
				"match_all": map[string]interface{}{},
			},
		}
	}

	var buf bytes.Buffer

	if err := json.NewEncoder(&buf).Encode(query); err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("error while parsing params"))
		return
	}

	resp, err := ESClient.Search(
		ESClient.Search.WithIndex(SearchIndex),
		ESClient.Search.WithBody(&buf),
	)

	defer resp.Body.Close()

	if err != nil || resp.IsError() {
		fmt.Println(err)
		fmt.Println(resp.Body.Close().Error())
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("error while searching data"))
		return
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("error while parsing response"))
		return
	}

	var data []User
	if hits, ok := result["hits"].(map[string]interface{}); ok {
		if hits2, ok := hits["hits"].([]interface{}); ok {
			for _, row := range hits2 {
				if row, ok := row.(map[string]interface{}); ok {
					if source, ok := row["_source"].(map[string]interface{}); ok {
						data = append(data, User{
							ID:        row["_id"].(string),
							FirstName: source["first_name"].(string),
							LastName:  source["last_name"].(string),
							Email:     source["email"].(string),
						})
					}
				}
			}
		}
	}

	json, err := json.Marshal(data)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("error while parsing response"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(json)
}

func PostUser(w http.ResponseWriter, r *http.Request) {
	var body User
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("error while parsing body"))
		return
	}

	data, err := json.Marshal(User{
		FirstName: body.FirstName,
		LastName:  body.LastName,
		Email:     body.Email,
	})

	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("error while parsing body 2"))
		return
	}

	req := esapi.IndexRequest{
		Index:   SearchIndex,
		Body:    bytes.NewReader(data),
		Refresh: "true",
	}

	resp, err := req.Do(context.Background(), ESClient)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("error while parsing body 2"))
		return
	}

	defer resp.Body.Close()
	log.Printf("Indexed document %s to index %s\n", resp.String(), SearchIndex)

	w.Write([]byte("Success index document"))
}

func UpdateUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		w.Write([]byte("ID Required"))
		return
	}

	var body User
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("error while parsing body"))
		return
	}

	data, err := json.Marshal(User{
		FirstName: body.FirstName,
		LastName:  body.LastName,
		Email:     body.Email,
	})

	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("error while parsing body 2"))
		return
	}

	req := esapi.UpdateRequest{
		Index:      SearchIndex,
		DocumentID: id,
		Body:       bytes.NewReader([]byte(fmt.Sprintf(`{"doc":%s}`, data))),
	}

	resp, err := req.Do(context.Background(), ESClient)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("error while parsing body 2"))
		return
	}

	defer resp.Body.Close()
	log.Printf("Updated document %s to index %s\n", resp.String(), SearchIndex)

	w.Write([]byte("Success update document"))
}

func DeleteUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		w.Write([]byte("ID Required"))
		return
	}

	req := esapi.DeleteRequest{
		Index:      SearchIndex,
		DocumentID: id,
	}

	resp, err := req.Do(context.Background(), ESClient)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("error while parsing body 2"))
		return
	}

	defer resp.Body.Close()
	log.Printf("Deleted document %s to index %s\n", resp.String(), SearchIndex)

	w.Write([]byte("Success delete document"))
}

func Info(w http.ResponseWriter, r *http.Request) {
	res, err := ESClient.Info()
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("error while parsing body"))
		return
	}

	data, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("error while parshig data"))
		return
	}
	defer res.Body.Close()
	w.Write([]byte(string(data)))
}

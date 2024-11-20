/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * open source MIT License, reproduced in the LICENSE file.
 */

package contacts

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"golang.org/x/time/rate"

	"github.com/clickonetwo/automations/dialpad/internal/storage"
)

var (
	dialpadApiRoot      = "https://dialpad.com/api/v2"
	dialPadListClient   = NewRLHTTPClient(1200, 60)
	dialPadUpdateClient = NewRLHTTPClient(100, 60)
	dialPadDeleteClient = NewRLHTTPClient(1200, 60)
)

type listEntry struct {
	Entry
	Id string `json:"id"`
}

type entryPage struct {
	Cursor string      `json:"cursor"`
	Items  []listEntry `json:"items"`
}

func ListContacts(accountId string) ([]Entry, error) {
	key := storage.GetConfig().DialpadApiKey
	baseUrl := fmt.Sprintf("%s/contacts?limit=200&apikey=%s", dialpadApiRoot, key)
	if accountId != "" {
		baseUrl = fmt.Sprintf("%s&accountId=%s", baseUrl, accountId)
	}
	results := make([]Entry, 0)
	cursor := ""
	for {
		url := baseUrl
		if cursor != "" {
			url = fmt.Sprintf("%s&cursor=%s", url, cursor)
		}
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			panic(err)
		}
		req.Header.Add("accept", "application/json")
		resp, err := dialPadListClient.Do(req)
		if err != nil {
			panic(err)
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		if resp.StatusCode != 200 {
			return results, fmt.Errorf("list contacts: status code: %d, body: %s", resp.StatusCode, body)
		}
		var result entryPage
		err = json.Unmarshal(body, &result)
		if err != nil {
			panic(err)
		}
		if len(result.Items) == 0 {
			break
		}
		for _, entry := range result.Items {
			entry.Entry.Uid = entry.Id
			results = append(results, entry.Entry)
		}
		cursor = result.Cursor
		if cursor == "" {
			break
		}
	}
	return results, nil
}

func UpdateContacts(entries []Entry) error {
	key := storage.GetConfig().DialpadApiKey
	url := fmt.Sprintf("%s/contacts?apikey=%s", dialpadApiRoot, key)
	for _, entry := range entries {
		body, err := json.Marshal(entry)
		if err != nil {
			panic(err)
		}
		req, _ := http.NewRequest("PUT", url, bytes.NewReader(body))
		req.Header.Add("accept", "application/json")
		req.Header.Add("content-type", "application/json")
		resp, err := dialPadUpdateClient.Do(req)
		if err != nil {
			panic(err)
		}
		body, _ = io.ReadAll(resp.Body)
		resp.Body.Close()
		if resp.StatusCode != 200 {
			return fmt.Errorf("contact: %v, status code: %d, body: %s", entry, resp.StatusCode, body)
		}
	}
	return nil
}

func DeleteContacts(entries []Entry) error {
	key := storage.GetConfig().DialpadApiKey
	for _, entry := range entries {
		url := fmt.Sprintf("%s/contacts/%s?apikey=%s", dialpadApiRoot, entry.Uid, key)
		req, _ := http.NewRequest("DELETE", url, nil)
		req.Header.Add("accept", "application/json")
		resp, err := dialPadDeleteClient.Do(req)
		if err != nil {
			panic(err)
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		if resp.StatusCode != 200 {
			return fmt.Errorf("contact id: %v, status code: %d, body: %s", entry.Uid, resp.StatusCode, body)
		}
	}
	return nil
}

// RLHTTPClient is rate-limited HTTP Client
//
// code taken from [this blogpost](https://medium.com/mflow/rate-limiting-in-golang-http-client-a22fba15861a)
type RLHTTPClient struct {
	client  *http.Client
	limiter *rate.Limiter
}

func NewRLHTTPClient(calls, seconds int) *RLHTTPClient {
	return &RLHTTPClient{
		client:  http.DefaultClient,
		limiter: rate.NewLimiter(rate.Every(time.Duration(seconds)*time.Second), calls),
	}
}

// Do dispatches the HTTP request to the network
func (c *RLHTTPClient) Do(req *http.Request) (*http.Response, error) {
	ctx := context.Background()
	err := c.limiter.Wait(ctx) // This is a blocking call. Honors the rate limit
	if err != nil {
		return nil, err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

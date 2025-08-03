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

	"golang.org/x/time/rate"

	"github.com/schollz/progressbar/v3"

	"github.com/clickonetwo/automations/dialpad/internal/storage"
)

var (
	DialpadApiRoot      = "https://dialpad.com/api/v2"
	DialPadListClient   = NewRLHTTPClient(12, 65)
	DialPadUpdateClient = NewRLHTTPClient(100, 60)
	DialPadDeleteClient = NewRLHTTPClient(20, 1)
)

type entryPage struct {
	Cursor string  `json:"cursor"`
	Items  []Entry `json:"items"`
}

func ListContacts(accountId string) (results []Entry, errs []error) {
	key := storage.GetConfig().DialpadApiKey
	baseUrl := fmt.Sprintf("%s/contacts?limit=100&apikey=%s", DialpadApiRoot, key)
	if accountId != "" {
		baseUrl = fmt.Sprintf("%s&accountId=%s", baseUrl, accountId)
	}
	cursor := ""
	bar := progressbar.Default(-1, "Downloading contacts")
	defer bar.Close()
	for {
		if len(errs) > 5 {
			panic(fmt.Errorf("too many errors fetching contacts: %v", errs))
		}
		url := baseUrl
		if cursor != "" {
			url = fmt.Sprintf("%s&cursor=%s", url, cursor)
		}
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			panic(err)
		}
		req.Header.Add("accept", "application/json")
		resp, err := DialPadListClient.Do(req)
		if err != nil {
			panic(err)
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		if resp.StatusCode != 200 {
			err := fmt.Errorf("status code: %d, body: %s", resp.StatusCode, body)
			//if strings.Contains(string(body), "\"limit\"") {
			//	time.Sleep(time.Second * 10)
			//}
			errs = append(errs, err)
			continue
		}
		var result entryPage
		err = json.Unmarshal(body, &result)
		if err != nil {
			panic(err)
		}
		if len(result.Items) == 0 {
			break
		}
		_ = bar.Add(len(result.Items))
		for _, entry := range result.Items {
			if entry.FullId == "" {
				err := fmt.Errorf("no contact ID: %v", entry)
				errs = append(errs, err)
				continue
			}
			entry.Uid, _ = ExtractUid(entry.FullId)
			results = append(results, entry)
		}
		cursor = result.Cursor
		if cursor == "" {
			break
		}
	}
	return
}

func UpdateContacts(entries []Entry) (errs []error) {
	key := storage.GetConfig().DialpadApiKey
	url := fmt.Sprintf("%s/contacts?apikey=%s", DialpadApiRoot, key)
	bar := progressbar.Default(int64(len(entries)))
	defer bar.Close()
	for _, entry := range entries {
		body, err := json.Marshal(entry)
		if err != nil {
			panic(err)
		}
		req, _ := http.NewRequest("PUT", url, bytes.NewReader(body))
		req.Header.Add("accept", "application/json")
		req.Header.Add("content-type", "application/json")
		resp, err := DialPadUpdateClient.Do(req)
		if err != nil {
			panic(err)
		}
		body, _ = io.ReadAll(resp.Body)
		resp.Body.Close()
		_ = bar.Add(1)
		if resp.StatusCode != 200 {
			err := fmt.Errorf("contact: %v, status code: %d, body: %s", entry, resp.StatusCode, body)
			errs = append(errs, err)
		}
	}
	return
}

func DeleteContacts(entries []Entry) (errs []error) {
	key := storage.GetConfig().DialpadApiKey
	bar := progressbar.Default(int64(len(entries)))
	defer bar.Close()
	for _, entry := range entries {
		if entry.FullId == "" {
			err := fmt.Errorf("no full ID for contact: %v", entry)
			errs = append(errs, err)
			continue
		}
		url := fmt.Sprintf("%s/contacts/%s?apikey=%s", DialpadApiRoot, entry.FullId, key)
		req, _ := http.NewRequest("DELETE", url, nil)
		req.Header.Add("accept", "application/json")
		resp, err := DialPadDeleteClient.Do(req)
		if err != nil {
			panic(err)
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		_ = bar.Add(1)
		if resp.StatusCode != 200 {
			err := fmt.Errorf("contact id: %v, status code: %d, body: %s", entry.Uid, resp.StatusCode, body)
			errs = append(errs, err)
		}
	}
	return
}

// RLHTTPClient is rate-limited HTTP Client
//
// code taken from [this blogpost](https://medium.com/mflow/rate-limiting-in-golang-http-client-a22fba15861a)
type RLHTTPClient struct {
	client  *http.Client
	limiter *rate.Limiter
}

func NewRLHTTPClient(calls, seconds int) *RLHTTPClient {
	limit := rate.Limit(calls) / rate.Limit(seconds)
	return &RLHTTPClient{
		client:  http.DefaultClient,
		limiter: rate.NewLimiter(limit, 1),
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

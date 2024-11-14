/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * open source MIT License, reproduced in the LICENSE file.
 */

package event

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/clickonetwo/automations/dialpad/internal/storage"
)

var (
	DialpadApiRootUrl = "https://dialpad.com/api/v2"
)

// EnsureWebHook guarantees that there's a registered Dialpad webhook
// with the given characteristics, and returns its hook ID.
//
// Hooks are defined by their hook URL and secret.  If an existing
// hook with matching characteristics exists, its ID is returned.
// Otherwise, a new hook is created and its ID is returned.
func EnsureWebHook(c context.Context, path, secret string) (string, error) {
	env := storage.GetConfig()
	hookUrl := env.HerokuHostUrl + path
	id, err := FindHook(c, hookUrl, secret)
	if err != nil {
		return "", err
	}
	if id != "" {
		return id, nil
	}
	id, err = CreateHook(c, hookUrl, secret)
	if err != nil {
		return "", err
	}
	return id, nil
}

// FindHook returns the ID of an existing webhook with the given characteristics, if any.
func FindHook(c context.Context, hookUrl, secret string) (string, error) {
	hooks, err := ListHooks(c)
	if err != nil {
		return "", err
	}
	for _, hook := range hooks {
		if hook.HookUrl != hookUrl {
			continue
		}
		if hook.Signature["secret"] != secret {
			continue
		}
		return hook.Id, nil
	}
	return "", nil
}

type HookDescriptor struct {
	HookUrl   string            `json:"hook_url"`
	Id        string            `json:"id"`
	Signature map[string]string `json:"signature"`
}

type hookListPage struct {
	Cursor string           `json:"cursor"`
	Items  []HookDescriptor `json:"items"`
}

// ListHooks retrieves descriptors for all the registered Dialpad web hooks.
//
// Right now, if there is more than one page of hooks, we only list the ones
// that appear on the first page.
//
// TODO: Make this work if there is more than one page of hooks
func ListHooks(c context.Context) ([]HookDescriptor, error) {
	apiUrl := fmt.Sprintf("%s/webhooks?apikey=%s", DialpadApiRootUrl, storage.GetConfig().DialpadApiKey)
	req, err := http.NewRequest(http.MethodGet, apiUrl, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", "application/json")
	resp, err := http.DefaultClient.Do(req.WithContext(c))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error getting hooks: %s", resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var page hookListPage
	if err := json.Unmarshal(body, &page); err != nil {
		return nil, err
	}
	return page.Items, nil
}

// CreateHook registers a dialpad webhook that calls this server.
//
// The created hook has the given path and secret. The id of the hook is returned.
func CreateHook(c context.Context, hookUrl, secret string) (string, error) {
	apiUrl := fmt.Sprintf("%s/webhooks?apikey=%s", DialpadApiRootUrl, storage.GetConfig().DialpadApiKey)
	params := map[string]string{"hook_url": hookUrl}
	if secret != "" {
		params["secret"] = secret
	}
	body, err := json.Marshal(params)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest(http.MethodPost, apiUrl, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req.WithContext(c))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("error creating hook: %s", resp.Status)
	}
	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var hook HookDescriptor
	if err = json.Unmarshal(body, &hook); err != nil {
		return "", err
	}
	return hook.Id, nil
}

// DeleteHook deletes the Dialpad webhook with the given id
func DeleteHook(c *gin.Context, id string) error {
	apiUrl := fmt.Sprintf("%s/webhooks/%s?apikey=%s", DialpadApiRootUrl, id, storage.GetConfig().DialpadApiKey)
	req, err := http.NewRequest(http.MethodDelete, apiUrl, nil)
	if err != nil {
		return err
	}
	req.Header.Add("Accept", "application/json")
	resp, err := http.DefaultClient.Do(req.WithContext(c.Request.Context()))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error deleting hook: %s", resp.Status)
	}
	return nil
}

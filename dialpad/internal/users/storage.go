/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * open source MIT License, reproduced in the LICENSE file.
 */

package users

import (
	"context"
	"encoding/json"

	"github.com/clickonetwo/automations/dialpad/internal/storage"
)

var (
	Admins  = UserList("admins")
	Readers = UserList("readers")
)

type UserList string

func (l UserList) StoragePrefix() string {
	return "users:"
}

func (l UserList) StorageId() string {
	return string(l)
}

func (l UserList) IdsEmails() (map[string]string, error) {
	val, err := storage.FetchString(context.Background(), l)
	if err != nil {
		return nil, err
	}
	var result map[string]string
	err = json.Unmarshal([]byte(val), &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (l UserList) SaveIdsEmails(users map[string]string) error {
	bytes, _ := json.Marshal(users)
	err := storage.StoreString(context.Background(), l, string(bytes))
	if err != nil {
		return err
	}
	return nil
}

/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * open source MIT License, reproduced in the LICENSE file.
 */

package storage

import (
	"fmt"

	"github.com/redis/go-redis/v9"
)

var (
	projectPrefix = "dialpad:"
	clientUrl     string
	client        *redis.Client
	keyPrefix     string
)

func GetDb() (*redis.Client, string) {
	config := GetConfig()
	if client != nil && clientUrl == config.DbUrl && keyPrefix == config.DbKeyPrefix {
		return client, keyPrefix
	}
	opts, err := redis.ParseURL(config.DbUrl)
	if err != nil {
		panic(fmt.Sprintf("invalid Redis url: %v", err))
	}
	clientUrl = config.DbUrl
	client = redis.NewClient(opts)
	keyPrefix = projectPrefix + config.DbKeyPrefix
	return client, keyPrefix
}

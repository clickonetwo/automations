/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * open source MIT License, reproduced in the LICENSE file.
 */

package storage

import (
	"fmt"
	"os"
	"strings"

	"github.com/dotenv-org/godotenvvault"
)

type Environment struct {
	Name                 string
	DbUrl                string
	DbKeyPrefix          string
	DialpadApiKey        string
	DialpadWebhookSecret string
}

var (
	testConfig = Environment{
		Name:                 "test",
		DbUrl:                "redis://",
		DbKeyPrefix:          "t:",
		DialpadApiKey:        "",
		DialpadWebhookSecret: "",
	}
	loadedConfig = testConfig
	configStack  []Environment
)

func GetConfig() Environment {
	return loadedConfig
}

func PushConfig(name string) error {
	if name == "" {
		return pushEnvConfig("")
	}
	if strings.HasPrefix(name, "t") {
		return pushTestConfig()
	}
	if strings.HasPrefix(name, "d") {
		return pushEnvConfig(".env")
	}
	if strings.HasPrefix(name, "s") {
		return pushEnvConfig(".env.staging")
	}
	if strings.HasPrefix(name, "p") {
		return pushEnvConfig(".env.production")
	}
	return fmt.Errorf("unknown environment: %s", name)
}

func PushAlteredConfig(env Environment) {
	configStack = append(configStack, loadedConfig)
	loadedConfig = env
}

func pushTestConfig() error {
	configStack = append(configStack, loadedConfig)
	loadedConfig = testConfig
	return nil
}

func pushEnvConfig(filename string) error {
	var d string
	var err error
	if filename == "" {
		if d, err = findEnvFile(".env.vault"); err == nil {
			if d == "" {
				err = godotenvvault.Overload()
			} else {
				var c string
				if c, err = os.Getwd(); err == nil {
					if err = os.Chdir(d); err == nil {
						err = godotenvvault.Overload()
						// if we fail to change back to the prior working directory, so be it.
						_ = os.Chdir(c)
					}
				}
			}
		}
	} else {
		if d, err = findEnvFile(filename); err == nil {
			err = godotenvvault.Overload(d + filename)
		}
	}
	if err != nil {
		return fmt.Errorf("error loading .env vars: %v", err)
	}
	configStack = append(configStack, loadedConfig)
	loadedConfig = Environment{
		Name:                 os.Getenv("ENVIRONMENT_NAME"),
		DbUrl:                os.Getenv("REDIS_URL"),
		DbKeyPrefix:          os.Getenv("DB_KEY_PREFIX"),
		DialpadApiKey:        os.Getenv("DIALPAD_API_KEY"),
		DialpadWebhookSecret: os.Getenv("DIALPAD_WEBHOOK_SECRET"),
	}
	return nil
}

func PopConfig() {
	if len(configStack) == 0 {
		return
	}
	loadedConfig = configStack[len(configStack)-1]
	configStack = configStack[:len(configStack)-1]
	return
}

func findEnvFile(name string) (string, error) {
	for i := range 5 {
		d := ""
		for range i {
			d += "../"
		}
		if _, err := os.Stat(d + name); err == nil {
			return d, nil
		}
	}
	return "", fmt.Errorf("no file %q found in path", name)
}

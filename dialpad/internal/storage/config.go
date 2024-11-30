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
	AgePublicKey         string
	AgeSecretKey         string
	AwsAccessKey         string
	AwsSecretKey         string
	AwsRegion            string
	AwsBucket            string
	AwsDialpadFolder     string
	DbUrl                string
	DbKeyPrefix          string
	DialpadApiKey        string
	DialpadWebhookSecret string
	HerokuHostUrl        string
	MasterAdminId        string
	MasterAdminEmail     string
}

//goland:noinspection SpellCheckingInspection
var (
	ciConfig = Environment{
		Name:                 "CI",
		AgePublicKey:         "age1mefeyagknwrhpepv6xkggfqwhgzj53rmuqmntkktf9y4mxavys3sns6prp",
		AgeSecretKey:         "AGE-SECRET-KEY-188SQQGWWCRCCYZ964VPDY5YGUCVL0QUEMEZ3TUV9AVJ2RY5G03QQE0PA38",
		AwsRegion:            "",
		AwsBucket:            "",
		AwsDialpadFolder:     "",
		DbUrl:                "redis://",
		DbKeyPrefix:          "c:",
		DialpadApiKey:        "",
		DialpadWebhookSecret: "",
		HerokuHostUrl:        "http://127.0.0.1:8080",
		MasterAdminId:        "dan@clickonetwo.io",
		MasterAdminEmail:     "dan@clickonetwo.io",
	}
	loadedConfig = ciConfig
	configStack  []Environment
)

func GetConfig() Environment {
	return loadedConfig
}

func PushConfig(name string) error {
	if name == "" {
		return pushEnvConfig("")
	}
	if strings.HasPrefix(name, "c") {
		return pushCiConfig()
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
	if strings.HasPrefix(name, "t") {
		return pushEnvConfig(".env.testing")
	}
	return fmt.Errorf("unknown environment: %s", name)
}

func PushAlteredConfig(env Environment) {
	configStack = append(configStack, loadedConfig)
	loadedConfig = env
}

func pushCiConfig() error {
	configStack = append(configStack, loadedConfig)
	loadedConfig = ciConfig
	return nil
}

func pushEnvConfig(filename string) error {
	var d string
	var err error
	if filename == "" {
		if d, err = FindEnvFile(".env.vault", true); err == nil {
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
		if d, err = FindEnvFile(filename, false); err == nil {
			err = godotenvvault.Overload(d + filename)
		}
	}
	if err != nil {
		return fmt.Errorf("error loading .env vars: %v", err)
	}
	configStack = append(configStack, loadedConfig)
	loadedConfig = Environment{
		Name:                 os.Getenv("ENVIRONMENT_NAME"),
		AgePublicKey:         os.Getenv("AGE_PUBLIC_KEY"),
		AgeSecretKey:         os.Getenv("AGE_SECRET_KEY"),
		AwsRegion:            os.Getenv("AWS_REGION"),
		AwsBucket:            os.Getenv("AWS_BUCKET"),
		AwsDialpadFolder:     os.Getenv("AWS_DIALPAD_FOLDER"),
		DbUrl:                os.Getenv("REDIS_URL"),
		DbKeyPrefix:          os.Getenv("DB_KEY_PREFIX"),
		DialpadApiKey:        os.Getenv("DIALPAD_API_KEY"),
		DialpadWebhookSecret: os.Getenv("DIALPAD_WEBHOOK_SECRET"),
		HerokuHostUrl:        os.Getenv("HEROKU_HOST_URL"),
		MasterAdminId:        os.Getenv("MASTER_ADMIN_ID"),
		MasterAdminEmail:     os.Getenv("MASTER_ADMIN_EMAIL"),
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

func FindEnvFile(name string, fallback bool) (string, error) {
	for i := range 5 {
		d := ""
		for range i {
			d += "../"
		}
		if _, err := os.Stat(d + name); err == nil {
			return d, nil
		}
		if fallback {
			if _, err := os.Stat(d + ".env"); err == nil {
				return d, nil
			}
		}
	}
	return "", fmt.Errorf("no file %q found in path", name)
}

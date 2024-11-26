/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * open source MIT License, reproduced in the LICENSE file.
 */

package middleware

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"reflect"
	"strings"

	"github.com/clickonetwo/automations/dialpad/internal/storage"
)

func init() {
	var tmp map[string]int64
	gob.Register(tmp)
}

type StatMap string

func (s StatMap) StoragePrefix() string {
	return "stat-map:"
}

func (s StatMap) StorageId() string {
	return string(s)
}

func (s StatMap) FetchAll() (map[string]any, error) {
	str, err := storage.FetchString(context.Background(), s)
	if err != nil {
		return nil, err
	}
	if str == "" {
		return make(map[string]any), nil
	}
	var val map[string]any
	dec := gob.NewDecoder(strings.NewReader(str))
	err = dec.Decode(&val)
	if err != nil {
		return nil, err
	}
	return val, nil
}

func (s StatMap) saveAll(stats map[string]any) error {
	if stats == nil {
		return storage.DeleteStorage(context.Background(), s)
	}
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(stats)
	if err != nil {
		return err
	}
	return storage.StoreString(context.Background(), s, buf.String())
}

func (s StatMap) Int64(name string) (int64, error) {
	stats, err := s.FetchAll()
	if err != nil {
		return 0, err
	}
	val, ok := stats[name]
	if !ok {
		return 0, nil
	}
	finalVal, ok := val.(int64)
	if !ok {
		return 0, fmt.Errorf("expected int64, got %s", reflect.TypeOf(val))
	}
	return finalVal, nil
}

func (s StatMap) SetInt64(name string, val int64) error {
	stats, err := s.FetchAll()
	if err != nil {
		return err
	}
	stats[name] = any(val)
	return s.saveAll(stats)
}

func (s StatMap) MapInt64(name string) (map[string]int64, error) {
	stats, err := s.FetchAll()
	if err != nil {
		return nil, err
	}
	val, ok := stats[name]
	if !ok {
		return make(map[string]int64), nil
	}
	finalVal, ok := val.(map[string]int64)
	if !ok {
		return nil, fmt.Errorf("expected map[string]int64, got %s", reflect.TypeOf(val))
	}
	if finalVal == nil {
		return make(map[string]int64), nil
	}
	return finalVal, nil
}

func (s StatMap) SetMapInt64(name string, val map[string]int64) error {
	stats, err := s.FetchAll()
	if err != nil {
		return err
	}
	stats[name] = any(val)
	return s.saveAll(stats)
}

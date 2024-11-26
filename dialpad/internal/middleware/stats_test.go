/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * open source MIT License, reproduced in the LICENSE file.
 */

package middleware

import (
	"context"
	"testing"

	"github.com/go-test/deep"

	"github.com/clickonetwo/automations/dialpad/internal/storage"
)

var statsTest = StatMap("test-statistics")

func TestFetchSaveFetchInt64(t *testing.T) {
	err := storage.DeleteStorage(context.Background(), statsTest)
	if err != nil {
		t.Fatal(err)
	}
	val, err := statsTest.Int64("test-int64")
	if err != nil {
		t.Fatal(err)
	}
	if val != 0 {
		t.Errorf("Expected 0, got %d", val)
	}
	err = statsTest.SetInt64("test-int64", 35)
	if err != nil {
		t.Fatal(err)
	}
	val, err = statsTest.Int64("test-int64")
	if err != nil {
		t.Error(err)
	}
	if val != 35 {

	}
}

func TestFetchSaveFetchMapInt64(t *testing.T) {
	err := storage.DeleteStorage(context.Background(), statsTest)
	val, err := statsTest.MapInt64("test-map-int64")
	if err != nil {
		t.Fatal(err)
	}
	if val != nil {
		t.Errorf("Expected nil, got %v", val)
	}
	m := map[string]int64{
		"fee": 0,
		"fie": 1,
		"foe": 2,
		"fum": 3,
	}
	err = statsTest.SetMapInt64("test-map-int64", m)
	if err != nil {
		t.Fatal(err)
	}
	val, err = statsTest.MapInt64("test-map-int64")
	if err != nil {
		t.Fatal(err)
	}
	if diff := deep.Equal(val, m); diff != nil {
		t.Error(diff)
	}
}

/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * open source MIT License, reproduced in the LICENSE file.
 */

package storage

import (
	"context"
	"os"
	"testing"
)

func TestS3GetPutBlob(t *testing.T) {
	err := PushConfig("testing")
	if err != nil {
		t.Fatal(err)
	}
	file, err := os.CreateTemp("", "test-get-blob-*.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(file.Name())
	defer file.Close()
	err = S3GetBlob(context.Background(), "sample.csv", file)
	if err != nil {
		t.Fatal(err)
	}
	stat, err := file.Stat()
	if err != nil {
		t.Fatal(err)
	}
	if count := stat.Size(); count != 17261 {
		t.Errorf("got %d bytes, want 17261 bytes", count)
	}
	_, err = file.Seek(0, 0)
	if err != nil {
		t.Fatal(err)
	}
	err = S3PutBlob(context.Background(), "sample.csv", file)
	if err != nil {
		t.Fatal(err)
	}
}

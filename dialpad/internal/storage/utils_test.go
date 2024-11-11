/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * open source MIT License, reproduced in the LICENSE file.
 */

package storage

import (
	"os"
	"testing"

	"github.com/google/uuid"
)

func TestDumpObjectsToPath(t *testing.T) {
	objects := ObjectMap{"string": []any{any("foobar")}, "int64": []any{any(int64(42))}}
	expected := `{
  "int64": [
    42
  ],
  "string": [
    "foobar"
  ]
}
`
	path := "/tmp/" + uuid.New().String() + ".json"
	if err := DumpObjectsToPath(objects, path); err != nil {
		t.Fatal(err)
	}
	if result, err := os.ReadFile(path); err != nil {
		t.Fatal(err)
	} else if expected != string(result) {
		t.Errorf("expected: %q, got: %q", expected, result)
	}
}

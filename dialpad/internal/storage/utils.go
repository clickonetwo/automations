/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * open source MIT License, reproduced in the LICENSE file.
 */

package storage

import (
	"encoding/json"
	"io"
	"os"
	"strings"
)

type ObjectMap map[string][]any

// DumpObjectsToPath serializes the entire map to the given filepath
func DumpObjectsToPath(what ObjectMap, where string) error {
	var stream io.Writer
	if where == "-" {
		// notest
		stream = os.Stdout
	} else {
		if !strings.HasSuffix(where, ".json") {
			where = where + ".json"
		}
		file, err := os.OpenFile(where, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o644)
		if err != nil {
			return err
		}
		defer file.Close()
		stream = file
	}
	DumpObjectsToStream(stream, what)
	return nil
}

// DumpObjectsToStream marshals the objects as JSON to the given stream
func DumpObjectsToStream(stream io.Writer, what ObjectMap) {
	encoder := json.NewEncoder(stream)
	encoder.SetIndent("", "  ")
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(what); err != nil {
		panic(err)
	}
}

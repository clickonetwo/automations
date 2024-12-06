/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * open source MIT License, reproduced in the LICENSE file.
 */

package migration

import (
	"encoding/json"
	"os"
)

func exportJsonToPath(obj any, path string) {
	f, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	e := json.NewEncoder(f)
	e.SetIndent("", "  ")
	e.SetEscapeHTML(false)
	err = e.Encode(obj)
	if err != nil {
		panic(err)
	}
}

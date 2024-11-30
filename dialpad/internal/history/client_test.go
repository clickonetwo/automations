/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * open source MIT License, reproduced in the LICENSE file.
 */

package history

import (
	"os"
	"testing"
)

func TestDownloadAndEncryptSmsReport(t *testing.T) {
	err := DownloadAndEncryptSmsReport("http://httpbin.org/get?this=is&a=test", "/tmp/test.json.age")
	if err != nil {
		t.Fatal(err)
	}
	_, err = os.Stat("/tmp/test.json.age")
	if err != nil {
		t.Fatal(err)
	}
}

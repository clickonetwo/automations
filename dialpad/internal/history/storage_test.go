/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * open source MIT License, reproduced in the LICENSE file.
 */

package history

import (
	"testing"

	"github.com/go-test/deep"

	"github.com/clickonetwo/automations/dialpad/internal/storage"
)

func TestUploadDowloadSmsHistory(t *testing.T) {
	err := storage.PushConfig("testing")
	if err != nil {
		t.Fatal(err)
	}
	uploaded, err := ImportEncryptedSmsEvents("../../local/texts-all.csv.age")
	if err != nil {
		t.Fatal(err)
	}
	err = UploadSmsHistory(uploaded)
	if err != nil {
		t.Fatal(err)
	}
	downloaded, err := DownloadSmsHistory()
	if err != nil {
		t.Fatal(err)
	}
	if diff := deep.Equal(uploaded, downloaded); diff != nil {
		t.Error(diff)
	}
}

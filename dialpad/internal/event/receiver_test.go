/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * open source MIT License, reproduced in the LICENSE file.
 */

package event

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-test/deep"
	"github.com/golang-jwt/jwt/v5"

	"github.com/clickonetwo/automations/dialpad/internal/auth"
	"github.com/clickonetwo/automations/dialpad/internal/middleware"
	"github.com/clickonetwo/automations/dialpad/internal/storage"
)

var (
	call = Call{
		CallId:      -1,
		DateStarted: time.Now().UnixMilli(),
		State:       "ringing",
	}
)

func TestCallStorableInterfaces(t *testing.T) {
	var c *Call = nil
	if c.StoragePrefix() != "call:" {
		t.Errorf("Calls have a non-'call:' prefix: %s", c.StoragePrefix())
	}
	if c.StorageId() != "" {
		t.Errorf("nil Call.StorageId() should return empty string")
	}
	if err := c.SetStorageId("test"); err == nil {
		t.Errorf("nil Call.SetStorageId() should error out")
	}
	if dup := c.Copy(); dup != nil {
		t.Errorf("nil Call.Copy() should return nil")
	}

	c = new(Call)
	*c = call
	if c.StorageId() != "-1" {
		t.Errorf("StorageId is wrong: %s != %s", c.StorageId(), "-1")
	}
	if err := c.SetStorageId("test"); err == nil {
		t.Errorf("Was able to set call storage id to non-numeric string")
	}
	if err := c.SetStorageId("30"); err != nil {
		t.Errorf("Failed to set storage id: %v", err)
	}
	if c.StorageId() != "30" {
		t.Errorf("StorageId is wrong: %s != %s", c.StorageId(), "30")
	}
	dup := c.Copy()
	if diff := deep.Equal(dup, c); diff != nil {
		t.Error(diff)
	}
	if dg, err := c.Downgrade(any(c)); err != nil {
		t.Error(err)
	} else if diff := deep.Equal(dg, c); diff != nil {
		t.Error(diff)
	}
	if dg, err := c.Downgrade(any(*c)); err != nil {
		t.Error(err)
	} else if diff := deep.Equal(dg, c); diff != nil {
		t.Error(diff)
	}
	if _, err := (*c).Downgrade(any(nil)); err == nil {
		t.Errorf("Call.Downgrade(nil) should error out")
	}
}

func TestReceiveWebhookPayloadNoSecret(t *testing.T) {
	r := middleware.CreateCoreEngine()
	r.POST("/hook", ReceiveCallWebhook)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/hook", marshalCallAsBody(t, call))
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Errorf("Wrong status code: %d", w.Code)
	}
	calls, err := storage.FetchRangeInterval(context.Background(), ReceivedCalls, -1, -1)
	if err != nil {
		t.Fatal(err)
	}
	if len(calls) != 1 {
		t.Errorf("Wrong number of calls: %d", len(calls))
	}
	if calls[0] != "-1" {
		t.Errorf("Wrong last call: %s", calls[0])
	}
}

func TestReceiveWebhookPayloadSecret(t *testing.T) {
	secret := auth.MakeNonce()
	key, _ := hex.DecodeString(secret)
	env := storage.GetConfig()
	env.DialpadWebhookSecret = secret
	storage.PushAlteredConfig(env)
	defer storage.PopConfig()
	token, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, marshalCallAsClaims(t, call)).SignedString(key)
	r := middleware.CreateCoreEngine()
	r.POST("/hook", ReceiveCallWebhook)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/hook", strings.NewReader(token))
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Errorf("wrong status code: %d", w.Code)
	}

}

func marshalCallAsBody(t *testing.T, p Call) io.Reader {
	b, err := json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	return bytes.NewReader(b)
}

func marshalCallAsClaims(t *testing.T, p Call) jwt.MapClaims {
	b, err := json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	var claims jwt.MapClaims
	err = json.Unmarshal(b, &claims)
	if err != nil {
		t.Fatal(err)
	}
	return claims
}

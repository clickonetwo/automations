/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * open source MIT License, reproduced in the LICENSE file.
 */

package event

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v5"

	"github.com/clickonetwo/automations/dialpad/internal/auth"
	"github.com/clickonetwo/automations/dialpad/internal/middleware"
	"github.com/clickonetwo/automations/dialpad/internal/storage"
)

var (
	call = Call{
		CallId:      "test1",
		DateStarted: 25,
		State:       "ringing",
	}
)

func TestReceiveWebhookPayloadNoSecret(t *testing.T) {
	r := middleware.CreateCoreEngine()
	r.POST("/hook", ReceiveCallWebhook)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/hook", marshalCallAsBody(t, call))
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Errorf("Wrong status code: %d", w.Code)
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
		t.Errorf("Wrong status code: %d", w.Code)
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

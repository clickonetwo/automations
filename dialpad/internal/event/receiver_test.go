/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * open source MIT License, reproduced in the LICENSE file.
 */

package event

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-test/deep"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"

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
	sampleHook = `{
		"call_id": 5314479320940544,
		"call_recording_ids": [],
		"callback_requested": null,
		"contact": {
		"email": "",
		"id": 6480210643730432,
		"name": "obscure user",
		"phone": "+15109109100",
		"type": "local"
		},
		"csat_score": null,
		"date_connected": null,
		"date_ended": null,
		"date_first_rang": 1731605955282,
		"date_queued": null,
		"date_rang": null,
		"date_started": 1731605954898,
		"direction": "inbound",
		"duration": null,
		"entry_point_call_id": null,
		"entry_point_target": {},
		"event_timestamp": 1731605955966,
		"external_number": "+15109109100",
		"group_id": null,
		"hold_time": null,
		"internal_number": "+15105105100",
		"is_transferred": false,
		"labels": [],
		"master_call_id": null,
		"mos_score": null,
		"operator_call_id": null,
		"proxy_target": {},
		"recording_details": [],
		"routing_breadcrumbs": [],
		"state": "ringing",
		"talk_time": null,
		"target": {
		"email": "obscured2@example.com",
		"id": 5991398486736896,
		"name": "obscure2 user",
		"office_id": 5527348325810176,
		"phone": "+15105105100",
		"type": "user"
		},
		"target_availability_status": "open",
		"total_duration": null,
		"transcription_text": null,
		"voicemail_link": null,
		"voicemail_recording_id": null,
		"was_recorded": false
    }`
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

func TestSaveFetchCall(t *testing.T) {
	ctx := context.Background()
	data := `{
		"call_id": 5314479320940544,
		"date_started": 1731605954898,
		"direction": "inbound",
		"external_number": "+15109260499",
		"internal_number": "+15105290106",
		"state": "ringing"
	}`
	var c Call
	if err := json.Unmarshal([]byte(data), &c); err != nil {
		t.Fatal(err)
	}
	if err := storage.SaveFields(ctx, &c); err != nil {
		t.Fatal(err)
	}

	c1 := Call{CallId: c.CallId}
	if err := storage.LoadFields(ctx, &c1); err != nil {
		t.Fatal(err)
	}
	if diff := deep.Equal(c, c1); diff != nil {
		t.Error(diff)
	}
	if err := storage.DeleteStorage(ctx, &c); err != nil {
		t.Fatal(err)
	}
}

func TestProcessCallWebhook(t *testing.T) {
	c, _ := middleware.CreateTestContext()
	if err := ProcessCallWebhook(c, json.RawMessage(sampleHook)); err != nil {
		t.Fatal(err)
	}
}

func TestReceiveWebhookPayloadNoSecret(t *testing.T) {
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	r := middleware.CreateCoreEngine(logger)
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
	env := storage.GetConfig()
	env.DialpadWebhookSecret = secret
	storage.PushAlteredConfig(env)
	defer storage.PopConfig()
	token, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, marshalCallAsClaims(t, call)).SignedString([]byte(secret))
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	r := middleware.CreateCoreEngine(logger)
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

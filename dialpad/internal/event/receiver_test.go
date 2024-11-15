/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * open source MIT License, reproduced in the LICENSE file.
 */

package event

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-test/deep"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"

	"github.com/clickonetwo/automations/dialpad/internal/auth"
	"github.com/clickonetwo/automations/dialpad/internal/middleware"
	"github.com/clickonetwo/automations/dialpad/internal/storage"
)

var (
	sampleCall = `{
		"call_id": 6421977457180672,
		"call_recording_ids": [],
		"callback_requested": null,
		"company_call_review_share_link": "https://dialpad.com/shared/call/YejKg43xavre2crkG6f3sk8nt5yV5ZNEN6IfcJVsselt",
		"contact": {
			"email": "",
			"id": 6131436978913280,
			"name": "Alameda, CA",
			"phone": "+15105105100",
			"type": "local"
		},
		"csat_score": null,
		"date_connected": null,
		"date_ended": 1731624048355,
		"date_first_rang": null,
		"date_queued": null,
		"date_rang": null,
		"date_started": 1731624039404,
		"direction": "inbound",
		"duration": 0,
		"entry_point_call_id": null,
		"entry_point_target": {},
		"event_timestamp": 1731624049994,
		"external_number": "+15102609745",
		"group_id": "Office:5527348325810176",
		"hold_time": 0,
		"internal_number": "+15106666687",
		"is_transferred": true,
		"labels": [],
		"master_call_id": null,
		"mos_score": 4.41,
		"operator_call_id": null,
		"proxy_target": {},
		"public_call_review_share_link": "https://dialpad.com/shared/call/cFuGfBym5Wod8vNX8KAItdOBXompUkNc8wIUgJhE85Pd",
		"recording_details": [],
		"routing_breadcrumbs": [],
		"state": "hangup",
		"talk_time": 0,
		"target": {
			"email": "",
			"id": 5527348325810176,
			"name": "Oasis Legal Services",
			"office_id": 5527348325810176,
			"phone": "+15106666687",
			"type": "office"
		},
		"target_availability_status": "open",
		"total_duration": 8951.029999999999,
		"transcription_text": null,
		"voicemail_link": null,
		"voicemail_recording_id": null,
		"was_recorded": false
	}`
	sampleSms = `{
		"admins": [
			{
			  "id": 6426782727880704,
			  "name": "Mari Brambila",
			  "phone_number": "(510) 972-9439",
			  "type": "user"
			}
		],
		"contact": {
			"id": 5914569395306496,
			"name": "Daniel Brotsky",
			"phone_number": "+15109260499"
		},
		"created_date": 1731632669183,
		"direction": "inbound",
		"event_timestamp": 1731632669601,
		"from_number": "+15109260499",
		"id": 6095823680520192,
		"is_internal": false,
		"message_delivery_result": null,
		"message_status": "pending",
		"mms": false,
		"mms_url": null,
		"sender_id": null,
		"target": {
			"id": 5527348325810176,
			"name": "Oasis Legal Services",
			"phone_number": "(510) 666-6687",
			"type": "office"
		},
		"text": "This is a test SMS - please ignore. Leanne can provide background information",
		"text_content": "This is a test SMS - please ignore. Leanne can provide background information",
		"to_number": [
		"+15106666687"
		]
    }`
)

func TestExtractWebhookPayload(t *testing.T) {
	c, _ := middleware.CreateTestContext()
	withoutSecret, err := extractWebhookPayload(c, json.RawMessage(sampleCall))
	if err != nil {
		t.Fatal(err)
	}
	secret := auth.MakeNonce()
	env := storage.GetConfig()
	env.DialpadWebhookSecret = secret
	storage.PushAlteredConfig(env)
	defer storage.PopConfig()
	claims := marshalPayloadAsClaims(t, json.RawMessage(sampleCall))
	token, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	withSecret, err := extractWebhookPayload(c, json.RawMessage(token))
	if err != nil {
		t.Fatal(err)
	}
	if diff := deep.Equal(withoutSecret, withSecret); diff != nil {
		t.Error(diff)
	}
}

func TestReceiveWebhookPayload(t *testing.T) {
	_ = storage.DeleteStorage(context.Background(), ActionHooks)
	_ = storage.DeleteStorage(context.Background(), IgnoreHooks)
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	r := middleware.CreateCoreEngine(logger)
	r.POST("/receive/:type", ReceiveWebhook)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/receive/call", strings.NewReader(sampleCall))
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Errorf("Wrong status code for call: %d", w.Code)
	}
	hooks, err := storage.FetchRangeInterval(context.Background(), IgnoreHooks, 0, -1)
	if err != nil {
		t.Fatal(err)
	}
	if len(hooks) != 1 {
		t.Errorf("Wrong number of ignore hooks: %d", len(hooks))
	}
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/receive/sms", strings.NewReader(sampleSms))
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Errorf("Wrong status code for SMS: %d", w.Code)
	}
	hooks, err = storage.FetchRangeInterval(context.Background(), ActionHooks, 0, -1)
	if err != nil {
		t.Fatal(err)
	}
	if len(hooks) != 1 {
		t.Errorf("Wrong number of action hooks: %d", len(hooks))
	}
}

func marshalPayloadAsClaims(t *testing.T, p json.RawMessage) jwt.MapClaims {
	var claims jwt.MapClaims
	err := json.Unmarshal(p, &claims)
	if err != nil {
		t.Fatal(err)
	}
	return claims
}

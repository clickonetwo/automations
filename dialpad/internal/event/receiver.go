/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * open source MIT License, reproduced in the LICENSE file.
 */

package event

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/clickonetwo/automations/dialpad/internal/auth"
	"github.com/clickonetwo/automations/dialpad/internal/middleware"
	"github.com/clickonetwo/automations/dialpad/internal/storage"
)

type jsObject map[string]interface{}

type HookSet string

func (e HookSet) StoragePrefix() string {
	return "hook-set:"
}

func (e HookSet) StorageId() string {
	return string(e)
}

var ActionHooks HookSet = "ActionHooks"
var IgnoreHooks HookSet = "IgnoreHooks"

func ReceiveWebhook(ctx *gin.Context) {
	defer ctx.Request.Body.Close()
	body, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		_ = ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	message, err := extractWebhookPayload(ctx, body)
	if err != nil {
		_ = ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}
	switch t := ctx.Param("type"); t {
	case "call":
		err = processCallWebhook(ctx, message)
	case "sms":
		err = processSmsWebhook(ctx, message)
	default:
		err = fmt.Errorf("unknown webhook type: %s", t)
	}
	if err != nil {
		_ = ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"status": "accepted"})
}

func extractWebhookPayload(ctx *gin.Context, body []byte) (jsObject, error) {
	var (
		message json.RawMessage
		err     error
	)
	env := storage.GetConfig()
	if env.DialpadWebhookSecret == "" {
		message = body
	} else {
		message, err = auth.ValidateDialpadJwt(ctx, string(body), env.DialpadWebhookSecret)
	}
	if err != nil {
		return nil, err
	}
	hook := make(map[string]interface{})
	if err := json.Unmarshal(message, &hook); err != nil {
		middleware.CtxLogS(ctx).Infow("Webhook parse error", "error", err, "payload", string(message))
		return nil, err
	}
	return hook, nil
}

func processCallWebhook(ctx *gin.Context, hook jsObject) error {
	targetSet := IgnoreHooks
	received := float64(time.Now().UnixMilli()) / 1000
	var state string
	if val, ok := hook["state"].(string); ok {
		state = val
	}
	var toNumber string
	if val, ok := hook["internal_number"].(string); ok {
		toNumber = val
	}
	switch state {
	case "voicemail", "voicemail_uploaded":
		middleware.CtxLogS(ctx).Infow(
			"Voicemail event",
			"state", state,
			"time", received,
			"contact", extractContact(hook, "contact"),
			"target", extractContact(hook, "target"),
			"to_number", toNumber,
			"url", hook["voicemail_link"],
		)
	case "connected":
		middleware.CtxLogS(ctx).Infow(
			"Call answered event",
			"state", state,
			"time", received,
			"contact", extractContact(hook, "contact"),
			"target", extractContact(hook, "target"),
			"to_number", toNumber,
		)
	default:
		middleware.CtxLogS(ctx).Infow("Ignoring call",
			"state", state,
			"time", received,
			"contact", extractContact(hook, "contact"),
			"target", extractContact(hook, "target"),
			"to_number", toNumber,
		)
		targetSet = IgnoreHooks
	}
	bytes, err := json.Marshal(hook)
	if err != nil {
		return err
	}
	err = storage.AddScoredMember(ctx.Request.Context(), targetSet, received, string(bytes))
	if err != nil {
		return err
	}
	return nil
}

func processSmsWebhook(ctx *gin.Context, hook jsObject) error {
	targetSet := IgnoreHooks
	received := float64(time.Now().UnixMilli()) / 1000
	var text string
	if val, ok := hook["text"].(string); ok {
		text = val
	}
	var toNumbers []string
	if val, ok := hook["to_number"].([]string); ok {
		toNumbers = val
	}
	middleware.CtxLogS(ctx).Infow(
		"Received SMS",
		"time", received,
		"contact", extractContact(hook, "contact"),
		"target", extractContact(hook, "target"),
		"to_numbers", toNumbers,
		"text", text,
	)
	bytes, err := json.Marshal(hook)
	if err != nil {
		return err
	}
	err = storage.AddScoredMember(ctx.Request.Context(), targetSet, received, string(bytes))
	if err != nil {
		return err
	}
	return nil
}

func extractContact(hook jsObject, label string) map[string]string {
	contact, ok := hook[label].(map[string]interface{})
	if !ok {
		return nil
	}
	m := map[string]string{"name": ""}
	if n, ok := contact["name"].(string); ok {
		m["name"] = n
	}
	if t, ok := contact["type"].(string); ok {
		m["type"] = t
	}
	if p, ok := contact["phone"].(string); ok {
		m["phone"] = p
	} else if p, ok = contact["phone_number"].(string); ok {
		m["phone"] = p
	}
	return m
}

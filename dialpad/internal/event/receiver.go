/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * open source MIT License, reproduced in the LICENSE file.
 */

package event

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/clickonetwo/automations/dialpad/internal/auth"
	"github.com/clickonetwo/automations/dialpad/internal/middleware"
	"github.com/clickonetwo/automations/dialpad/internal/storage"
)

type Call struct {
	CallId           string  `json:"call_id"`
	Contact          Contact `json:"contact,omitempty"`
	DateRang         int64   `json:"date_rang,omitempty"`
	DateStarted      int64   `json:"date_started"`
	Direction        string  `json:"direction"`
	Duration         int64   `json:"duration,omitempty"`
	EntryPointCallId string  `json:"entry_point_call_id,omitempty"`
	EntryPointTarget string  `json:"entry_point_target,omitempty"`
	ExternalNumber   string  `json:"external_number"`
	InternalNumber   string  `json:"internal_number"`
	State            string  `json:"state"`
}

type Contact struct {
	Phone string `json:"phone"`
	Type  string `json:"type"`
	Id    string `json:"id"`
	Email string `json:"email,omitempty"`
	Name  string `json:"name,omitempty"`
}

func ReceiveCallWebhook(ctx *gin.Context) {
	var message json.RawMessage
	body, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		_ = ctx.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	env := storage.GetConfig()
	if env.DialpadWebhookSecret == "" {
		message = body
	} else {
		message, err = auth.ValidateDialpadJwt(ctx, string(body), env.DialpadWebhookSecret)
		if err != nil {
			_ = ctx.AbortWithError(http.StatusBadRequest, err)
			return
		}
	}
	if err = ProcessCallWebhook(ctx, message); err != nil {
		_ = ctx.AbortWithError(http.StatusBadRequest, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"status": "accepted"})
}

func ProcessCallWebhook(ctx *gin.Context, message json.RawMessage) error {
	middleware.CtxLogS(ctx).Infow("Received dialpad call payload", "payload", string(message))
	var p Call
	if err := json.Unmarshal(message, &p); err != nil {
		return err
	}
	middleware.CtxLogS(ctx).Infow("Parsed dialpad call", "call", p)
	return nil
}

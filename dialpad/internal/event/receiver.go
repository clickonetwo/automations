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
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/clickonetwo/automations/dialpad/internal/auth"
	"github.com/clickonetwo/automations/dialpad/internal/middleware"
	"github.com/clickonetwo/automations/dialpad/internal/storage"
)

type CallSet string

func (e CallSet) StoragePrefix() string {
	return "call-set:"
}

func (e CallSet) StorageId() string {
	return string(e)
}

type Call struct {
	// fields with call data
	CallId         int64  `redis:"callId" json:"call_id"`
	DateRang       int64  `redis:"dateRang" json:"date_rang,omitempty"`
	DateStarted    int64  `redis:"dateStarted" json:"date_started"`
	Direction      string `redis:"direction" json:"direction"`
	Duration       int64  `redis:"duration" json:"duration,omitempty"`
	ExternalNumber string `redis:"externalNumber" json:"external_number"`
	InternalNumber string `redis:"internalNumber" json:"internal_number"`
	State          string `redis:"state" json:"state"`
	// fields with webhook data
	DateReceived int64  `redis:"dateReceived" json:"-"`
	AsReceived   string `redis:"asReceived" json:"-"`
}

func (c *Call) StoragePrefix() string {
	return "call:"
}

func (c *Call) StorageId() string {
	if c == nil {
		return ""
	}
	return fmt.Sprintf("%d", c.CallId)
}

func (c *Call) SetStorageId(id string) error {
	if c == nil {
		return fmt.Errorf("can't set storage id for nil Call")
	}
	val, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return err
	}
	c.CallId = val
	return nil
}

func (c *Call) Copy() storage.StructPointer {
	if c == nil {
		return nil
	}
	cpy := *c
	return &cpy
}

func (c *Call) Downgrade(a any) (storage.StructPointer, error) {
	if ac, ok := a.(Call); ok {
		return &ac, nil
	}
	if ac, ok := a.(*Call); ok {
		return ac, nil
	}
	return nil, fmt.Errorf("not a Call: %#v", a)
}

var ReceivedCalls CallSet = "ReceivedCalls"

func ReceiveCallWebhook(ctx *gin.Context) {
	var message json.RawMessage
	body, err := io.ReadAll(ctx.Request.Body)
	_ = ctx.Request.Body.Close()
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
	p := Call{DateReceived: time.Now().UnixMilli(), AsReceived: string(message)}
	if err := json.Unmarshal(message, &p); err != nil {
		middleware.CtxLogS(ctx).Infow("Call parse error", "error", err, "payload", string(message))
	} else {
		middleware.CtxLogS(ctx).Infow("Parsed call", "call", p)
	}
	if err := storage.SaveFields(ctx.Request.Context(), &p); err != nil {
		return err
	}
	if err := storage.AddScoredMember(ctx.Request.Context(), ReceivedCalls, p.DateReceived, p.StorageId()); err != nil {
		return err
	}
	return nil
}

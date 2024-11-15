/*
 * Copyright 2024 Daniel C. Brotsky. All rights reserved.
 * All the copyrighted work in this repository is licensed under the
 * open source MIT License, reproduced in the LICENSE file.
 */

package cmd

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/clickonetwo/automations/dialpad/internal/event"
	"github.com/clickonetwo/automations/dialpad/internal/middleware"
	"github.com/clickonetwo/automations/dialpad/internal/storage"
)

// receiveCmd represents the receive command
var receiveCmd = &cobra.Command{
	Use:   "receive",
	Short: "Receive webhook payloads from Dialpad",
	Long: `This command runs a server process that receives webhook payloads from Dialpad.
The registration of the webhook endpoint that receives the calls is done by this process.
The registration of the subscriptions that generate the webhooks must be done separately.

The server, in addition to serving the webhook endpoint, serves a /status endpoint
which reports the webhook ID in use so it can be used to create subscriptions.`,
	Run: func(cmd *cobra.Command, args []string) {
		envName, err := cmd.Flags().GetString("env")
		if err != nil {
			panic(err)
		}
		receive(envName)
	},
}

func init() {
	rootCmd.AddCommand(receiveCmd)
	receiveCmd.Flags().StringP("env", "e", "", "processing environment")
}

func receive(envName string) {
	startTime := time.Now()
	_ = storage.PushConfig(envName)
	defer storage.PopConfig()
	config := storage.GetConfig()
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()
	callId, err := event.EnsureWebHook(context.Background(), "/receive/call", config.DialpadWebhookSecret)
	if err != nil {
		panic(err)
	}
	smsId, err := event.EnsureWebHook(context.Background(), "/receive/sms", config.DialpadWebhookSecret)
	if err != nil {
		panic(err)
	}
	logger.Info("Registered webhooks at startup", zap.String("calls", callId), zap.String("sms", smsId))
	if config.Name == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := middleware.CreateCoreEngine(logger)
	r.POST("/receive/:type", event.ReceiveWebhook)
	r.GET("/status", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "receiver running",
			"env":       config.Name,
			"started":   startTime.String(),
			"time":      time.Since(startTime).String(),
			"call_hook": callId,
			"sms_hook":  smsId,
		})
	})
	port, found := os.LookupEnv("PORT")
	if !found {
		port = "8080"
	}
	address := "0.0.0.0:" + port
	if config.Name == "development" {
		address = "127.0.0.1:" + port
	}
	logger.Info(
		"Listening on address",
		zap.String("endpoint", config.HerokuHostUrl),
		zap.String("address", address),
	)
	if err := r.Run(address); err != nil {
		panic(err)
	}
}
